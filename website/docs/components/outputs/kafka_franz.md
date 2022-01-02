---
title: kafka_franz
type: output
status: experimental
categories: ["Services"]
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the contents of:
     lib/output/kafka_franz.go
-->

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

:::caution EXPERIMENTAL
This component is experimental and therefore subject to change or removal outside of major version releases.
:::
An alternative Kafka output using the [Franz Kafka client library](https://github.com/twmb/franz-go).

Introduced in version 3.61.0.


<Tabs defaultValue="common" values={[
  { label: 'Common', value: 'common', },
  { label: 'Advanced', value: 'advanced', },
]}>

<TabItem value="common">

```yaml
# Common config fields, showing default values
output:
  label: ""
  kafka_franz:
    seed_brokers: []
    topic: ""
    key: ""
    metadata:
      include_prefixes: []
      include_patterns: []
    max_in_flight: 10
    batching:
      count: 0
      byte_size: 0
      period: ""
      check: ""
```

</TabItem>
<TabItem value="advanced">

```yaml
# All config fields, showing default values
output:
  label: ""
  kafka_franz:
    seed_brokers: []
    topic: ""
    key: ""
    partitioner: ""
    metadata:
      include_prefixes: []
      include_patterns: []
    max_in_flight: 10
    batching:
      count: 0
      byte_size: 0
      period: ""
      check: ""
      processors: []
    max_message_bytes: 1MB
    compression: ""
    tls:
      enabled: false
      skip_cert_verify: false
      enable_renegotiation: false
      root_cas: ""
      root_cas_file: ""
      client_certs: []
    sasl: []
```

</TabItem>
</Tabs>

Consumes one or more topics by balancing the partitions across any other connected clients with the same consumer group.

This input is new and experimental, and the existing `kafka` input is not going anywhere, but here's some reasons why it might be worth trying this one out:

- You like shiny new stuff
- You are exeriencing issues with the existing `kafka` input
- Someone told you to


## Fields

### `seed_brokers`

A list of broker addresses to connect to in order to establish connections. If an item of the list contains commas it will be expanded into multiple addresses.


Type: `array`  

```yaml
# Examples

seed_brokers:
  - localhost:9092

seed_brokers:
  - foo:9092
  - bar:9092

seed_brokers:
  - foo:9092,bar:9092
```

### `topic`

A topic to write messages to.
This field supports [interpolation functions](/docs/configuration/interpolation#bloblang-queries).


Type: `string`  

### `key`

An optional key to populate for each message.
This field supports [interpolation functions](/docs/configuration/interpolation#bloblang-queries).


Type: `string`  

### `partitioner`

Override the default murmur2 hashing partitioner.


Type: `string`  

| Option | Summary |
|---|---|
| `least_backup` | Chooses the least backed up partition (the partition with the fewest amount of buffered records). Partitions are selected per batch. |
| `round_robin` | Round-robin's messages through all available partitions. This algorithm has lower throughput and causes higher CPU load on brokers, but can be useful if you want to ensure an even distribution of records to partitions. |


### `metadata`

Determine which (if any) metadata values should be added to messages as headers.


Type: `object`  

### `metadata.include_prefixes`

Provide a list of explicit metadata key prefixes to match against.


Type: `array`  

```yaml
# Examples

include_prefixes:
  - foo_
  - bar_

include_prefixes:
  - kafka_

include_prefixes:
  - content-
```

### `metadata.include_patterns`

Provide a list of explicit metadata key regular expression (re2) patterns to match against.


Type: `array`  

```yaml
# Examples

include_patterns:
  - .*

include_patterns:
  - _timestamp_unix$
```

### `max_in_flight`

The maximum number of batches to be sending in parallel at any given time.


Type: `int`  
Default: `10`  

### `batching`

Allows you to configure a [batching policy](/docs/configuration/batching).


Type: `object`  

```yaml
# Examples

batching:
  byte_size: 5000
  count: 0
  period: 1s

batching:
  count: 10
  period: 1s

batching:
  check: this.contains("END BATCH")
  count: 0
  period: 1m
```

### `batching.count`

A number of messages at which the batch should be flushed. If `0` disables count based batching.


Type: `int`  
Default: `0`  

### `batching.byte_size`

An amount of bytes at which the batch should be flushed. If `0` disables size based batching.


Type: `int`  
Default: `0`  

### `batching.period`

A period in which an incomplete batch should be flushed regardless of its size.


Type: `string`  
Default: `""`  

```yaml
# Examples

period: 1s

period: 1m

period: 500ms
```

### `batching.check`

A [Bloblang query](/docs/guides/bloblang/about/) that should return a boolean value indicating whether a message should end a batch.


Type: `string`  
Default: `""`  

```yaml
# Examples

check: this.type == "end_of_transaction"
```

### `batching.processors`

A list of [processors](/docs/components/processors/about) to apply to a batch as it is flushed. This allows you to aggregate and archive the batch however you see fit. Please note that all resulting messages are flushed as a single batch, therefore splitting the batch into smaller batches using these processors is a no-op.


Type: `array`  

```yaml
# Examples

processors:
  - archive:
      format: lines

processors:
  - archive:
      format: json_array

processors:
  - merge_json: {}
```

### `max_message_bytes`

The maximum space in bytes than an individual message may take, messages larger than this value will be rejected. This field corresponds to Kafka's `max.message.bytes`.


Type: `string`  
Default: `"1MB"`  

```yaml
# Examples

max_message_bytes: 100MB

max_message_bytes: 50mib
```

### `compression`

Optionally set an explicit compression type. The default preference is to use snappy when the broker supports it, and fall back to none if not.


Type: `string`  
Options: `lz4`, `snappy`, `gzip`, `none`, `zstd`.

### `tls`

Custom TLS settings can be used to override system defaults.


Type: `object`  

### `tls.enabled`

Whether custom TLS settings are enabled.


Type: `bool`  
Default: `false`  

### `tls.skip_cert_verify`

Whether to skip server side certificate verification.


Type: `bool`  
Default: `false`  

### `tls.enable_renegotiation`

Whether to allow the remote server to repeatedly request renegotiation. Enable this option if you're seeing the error message `local error: tls: no renegotiation`.


Type: `bool`  
Default: `false`  
Requires version 3.45.0 or newer  

### `tls.root_cas`

An optional root certificate authority to use. This is a string, representing a certificate chain from the parent trusted root certificate, to possible intermediate signing certificates, to the host certificate.


Type: `string`  
Default: `""`  

```yaml
# Examples

root_cas: |-
  -----BEGIN CERTIFICATE-----
  ...
  -----END CERTIFICATE-----
```

### `tls.root_cas_file`

An optional path of a root certificate authority file to use. This is a file, often with a .pem extension, containing a certificate chain from the parent trusted root certificate, to possible intermediate signing certificates, to the host certificate.


Type: `string`  
Default: `""`  

```yaml
# Examples

root_cas_file: ./root_cas.pem
```

### `tls.client_certs`

A list of client certificates to use. For each certificate either the fields `cert` and `key`, or `cert_file` and `key_file` should be specified, but not both.


Type: `array`  

```yaml
# Examples

client_certs:
  - cert: foo
    key: bar

client_certs:
  - cert_file: ./example.pem
    key_file: ./example.key
```

### `tls.client_certs[].cert`

A plain text certificate to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].key`

A plain text certificate key to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].cert_file`

The path to a certificate to use.


Type: `string`  
Default: `""`  

### `tls.client_certs[].key_file`

The path of a certificate key to use.


Type: `string`  
Default: `""`  

### `sasl`

Specify one or more methods of SASL authentication. SASL is tried in order; if the broker supports the first mechanism, all connections will use that mechanism. If the first mechanism fails, the client will pick the first supported mechanism. If the broker does not support any client mechanisms, connections will fail.


Type: `array`  

```yaml
# Examples

sasl:
  - mechanism: SCRAM-SHA-512
    password: bar
    username: foo
```

### `sasl[].mechanism`

The SASL mechanism to use.


Type: `string`  

| Option | Summary |
|---|---|
| `OAUTHBEARER` | OAuth Bearer based authentication. |
| `PLAIN` | Plain text authentication. |
| `SCRAM-SHA-256` | SCRAM based authentication as specified in RFC5802. |
| `SCRAM-SHA-512` | SCRAM based authentication as specified in RFC5802. |


### `sasl[].username`

A username to provide for PLAIN or SCRAM-* authentication.


Type: `string`  
Default: `""`  

### `sasl[].password`

A password to provide for PLAIN or SCRAM-* authentication.


Type: `string`  
Default: `""`  

### `sasl[].token`

The token to use for a single session's OAUTHBEARER authentication.


Type: `string`  
Default: `""`  

### `sasl[].extensions`

Key/value pairs to add to OAUTHBEARER authentication requests.


Type: `object`  

