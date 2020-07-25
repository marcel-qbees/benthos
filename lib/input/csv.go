package input

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Jeffail/benthos/v3/lib/input/reader"
	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/message"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/types"
	"github.com/Jeffail/benthos/v3/lib/x/docs"
)

//------------------------------------------------------------------------------

func init() {
	Constructors[TypeCSVFile] = TypeSpec{
		constructor: NewCSVFile,
		Summary: `
BETA: This component is experimental and therefore subject to change outside of
major version releases.

Reads one or more CSV files as structured records.`,
		FieldSpecs: docs.FieldSpecs{
			docs.FieldCommon("paths", "A list of file paths to read from. Each file will be read sequentially until the list is exhausted, at which point the input will close."),
			docs.FieldCommon("delimiter", `The delimiter to use for splitting values in each record, must be a single character.`),
		},
	}
}

//------------------------------------------------------------------------------

// CSVFileConfig contains configuration values for the CSVFile input type.
type CSVFileConfig struct {
	Paths []string `json:"paths" yaml:"paths"`
	Delim string   `json:"delimiter" yaml:"delimiter"`
}

// NewCSVFileConfig creates a new CSVFileConfig with default values.
func NewCSVFileConfig() CSVFileConfig {
	return CSVFileConfig{
		Paths: []string{},
		Delim: ",",
	}
}

//------------------------------------------------------------------------------

// NewCSVFile creates a new CSV file input type.
func NewCSVFile(conf Config, mgr types.Manager, log log.Modular, stats metrics.Type) (Type, error) {
	delimRunes := []rune(conf.CSVFile.Delim)
	if len(delimRunes) != 1 {
		return nil, errors.New("delimiter value must be exactly one character")
	}

	comma := delimRunes[0]

	pathsRemaining := conf.CSVFile.Paths
	if len(pathsRemaining) == 0 {
		return nil, errors.New("requires at least one input file path")
	}

	rdr, err := newCSVReader(
		func(context.Context) (io.Reader, error) {
			if len(pathsRemaining) == 0 {
				return nil, io.EOF
			}

			path := pathsRemaining[0]
			pathsRemaining = pathsRemaining[1:]

			return os.Open(path)
		},
		func(context.Context) {},
		optCSVSetComma(comma),
	)
	if err != nil {
		return nil, err
	}

	return NewAsyncReader(TypeFile, true, reader.NewAsyncPreserver(rdr), log, stats)
}

//------------------------------------------------------------------------------

// csvReader is an reader. implementation that consumes an io.Reader and parses
// it as a CSV file.
type csvReader struct {
	handleCtor func(ctx context.Context) (io.Reader, error)
	onClose    func(ctx context.Context)

	mut        sync.Mutex
	handle     io.Reader
	shutdownFn func()
	errChan    chan error
	msgChan    chan types.Message

	expectHeaders bool
	comma         rune
	strict        bool
}

// NewCSV creates a new reader input type able to create a feed of line
// delimited CSV records from an io.Reader.
//
// Callers must provide a constructor function for the target io.Reader, which
// is called on start up and again each time a reader is exhausted. If the
// constructor is called but there is no more content to create a Reader for
// then the error `io.EOF` should be returned and the CSV will close.
//
// Callers must also provide an onClose function, which will be called if the
// CSV has been instructed to shut down. This function should unblock any
// blocked Read calls.
func newCSVReader(
	handleCtor func(ctx context.Context) (io.Reader, error),
	onClose func(ctx context.Context),
	options ...func(r *csvReader),
) (*csvReader, error) {
	r := csvReader{
		handleCtor:    handleCtor,
		onClose:       onClose,
		comma:         ',',
		expectHeaders: true,
		strict:        false,
	}

	for _, opt := range options {
		opt(&r)
	}

	r.shutdownFn = func() {}
	return &r, nil
}

//------------------------------------------------------------------------------

// OptCSVSetComma is a option func that sets the comma character (default ',')
// to be used to divide records.
func optCSVSetComma(comma rune) func(r *csvReader) {
	return func(r *csvReader) {
		r.comma = comma
	}
}

// OptCSVSetExpectHeaders is a option func that determines whether the first
// record from the CSV input outlines the names of columns.
func optCSVSetExpectHeaders(expect bool) func(r *csvReader) {
	return func(r *csvReader) {
		r.expectHeaders = expect
	}
}

// OptCSVSetStrict is a option func that determines whether records with
// misaligned numbers of fields should be rejected.
func optCSVSetStrict(strict bool) func(r *csvReader) {
	return func(r *csvReader) {
		r.strict = strict
	}
}

//------------------------------------------------------------------------------

func (r *csvReader) closeHandle() {
	if r.handle != nil {
		if closer, ok := r.handle.(io.ReadCloser); ok {
			closer.Close()
		}
		r.handle = nil
	}
	r.shutdownFn()
}

//------------------------------------------------------------------------------

// ConnectWithContext attempts to establish a new scanner for an io.Reader.
func (r *csvReader) ConnectWithContext(ctx context.Context) error {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.closeHandle()

	handle, err := r.handleCtor(ctx)
	if err != nil {
		if err == io.EOF {
			return types.ErrTypeClosed
		}
		return err
	}

	scanner := csv.NewReader(handle)
	scanner.Comma = r.comma
	scanner.ReuseRecord = true

	scannerCtx, shutdownFn := context.WithCancel(context.Background())
	msgChan := make(chan types.Message)
	errChan := make(chan error)

	go func() {
		defer func() {
			shutdownFn()
			close(errChan)
			close(msgChan)
		}()

		var headers []string

	recordLoop:
		for {
			records, err := scanner.Read()
			if err != nil && (r.strict || len(records) == 0) {
				if err == io.EOF {
					break recordLoop
				}
				select {
				case errChan <- err:
				case <-scannerCtx.Done():
					return
				}
				continue recordLoop
			}

			if r.expectHeaders && headers == nil {
				headers = make([]string, 0, len(records))
				for _, r := range records {
					headers = append(headers, r)
				}
				continue recordLoop
			}

			part := message.NewPart(nil)

			var structured interface{}
			if len(headers) == 0 || len(headers) < len(records) {
				slice := make([]interface{}, 0, len(records))
				for _, r := range records {
					slice = append(slice, r)
				}
				structured = slice
			} else {
				obj := make(map[string]interface{}, len(records))
				for i, r := range records {
					obj[headers[i]] = r
				}
				structured = obj
			}

			if err = part.SetJSON(structured); err != nil {
				select {
				case errChan <- err:
				case <-scannerCtx.Done():
					return
				}
				continue recordLoop
			}

			msg := message.New(nil)
			msg.Append(part)
			select {
			case msgChan <- msg:
			case <-scannerCtx.Done():
				return
			}
		}
	}()

	r.handle = handle
	r.msgChan = msgChan
	r.errChan = errChan
	r.shutdownFn = shutdownFn
	return nil
}

// ReadWithContext attempts to read a new line from the io.Reader.
func (r *csvReader) ReadWithContext(ctx context.Context) (types.Message, reader.AsyncAckFn, error) {
	r.mut.Lock()
	msgChan := r.msgChan
	errChan := r.errChan
	r.mut.Unlock()

	select {
	case msg, open := <-msgChan:
		if !open {
			return nil, nil, types.ErrNotConnected
		}
		return msg, func(context.Context, types.Response) error { return nil }, nil
	case err, open := <-errChan:
		if !open {
			return nil, nil, types.ErrNotConnected
		}
		return nil, nil, err
	case <-ctx.Done():
	}
	return nil, nil, types.ErrTimeout
}

// CloseAsync shuts down the reader input and stops processing requests.
func (r *csvReader) CloseAsync() {
	go func() {
		r.mut.Lock()
		r.onClose(context.Background())
		r.closeHandle()
		r.mut.Unlock()
	}()
}

// WaitForClose blocks until the reader input has closed down.
func (r *csvReader) WaitForClose(timeout time.Duration) error {
	return nil
}

//------------------------------------------------------------------------------