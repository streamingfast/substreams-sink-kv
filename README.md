# Substreams sink key-value store

## Description

`substreams-sink-kv` is a tool that allows developers to pipe data extracted from a blockchain into a key-value store and expose it through Connect-Web protocol (compatible with GRPC)

## Prerequisites

- Rust installation and compiler
- Cloned `substreams-sink-kv` repository
- A Substreams module prepared for a kv-sink
- Knowledge of blockchain and substreams development

## Installation

### From source

* This requires the [Go compiler](https://go.dev/)
* Clone this repository, then run `go install -v ./cmd/substreams-sink-kv`
* Check your installation by running `substreams-sink-kv --version`

> **Note** the binary file will be installed in your *GO_PATH*, usually `$HOME/go/bin`. Ensure that this folder is included in your *PATH* environment variable.

### From a pre-built binary release

* Download the release from here: https://github.com/streamingfast/substreams-sink-kv/releases
* Extract the `substreams-sink-kv` from the tarball into a folder available in your PATH.

## Running with an example substreams

> **Note** To connect to substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

```bash
substreams-sink-kv \
  run \
  "badger3://$(pwd)/badger_data.db" \
  mainnet.eth.streamingfast.io:443 \
  https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg \
  kv_out
```

You should see output similar to this one:
```bash
2023-01-12T10:08:31.803-0500 INFO (sink-kv) starting prometheus metrics server {"listen_addr": "localhost:9102"}
2023-01-12T10:08:31.803-0500 INFO (sink-kv) sink to kv {"dsn": "badger3:///Users/stepd/repos/substreams-sink-kv/badger_data.db", "endpoint": "mainnet.eth.streamingfast.io:443", "manifest_path": "https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg", "output_module_name": "kv_out", "block_range": ""}
2023-01-12T10:08:31.803-0500 INFO (sink-kv) starting pprof server {"listen_addr": "localhost:6060"}
2023-01-12T10:08:31.826-0500 INFO (sink-kv) reading substreams manifest {"manifest_path": "https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg"}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) validating output store {"output_store": "kv_out"}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) resolved block range {"start_block": 0, "stop_block": 0}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) starting to listen on {"addr": "localhost:8000"}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) starting stats service {"runs_each": "2s"}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) no block data buffer provided. since undo steps are possible, using default buffer size {"size": 12}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) starting stats service {"runs_each": "2s"}
2023-01-12T10:08:32.186-0500 INFO (sink-kv) ready, waiting for signal to quit
2023-01-12T10:08:32.186-0500 INFO (sink-kv) launching server {"listen_addr": "localhost:8000"}
2023-01-12T10:08:32.187-0500 INFO (sink-kv) serving plaintext {"listen_addr": "localhost:8000"}
2023-01-12T10:08:32.278-0500 INFO (sink-kv) session init {"trace_id": "a3c59bd7992c433402b70f9541565d2d"}
2023-01-12T10:08:34.186-0500 INFO (sink-kv) substreams sink stats {"db_flush_rate": "10.500 flush/s (21 total)", "data_msg_rate": "0.000 msg/s (0 total)", "progress_msg_rate": "0.000 msg/s (0 total)", "block_rate": "0.000 blocks/s (0 total)", "flushed_entries": 0, "last_block": "None"}
2023-01-12T10:08:34.186-0500 INFO (sink-kv) substreams sink stats {"progress_msg_rate": "16551.500 msg/s (33103 total)", "block_rate": "10941.500 blocks/s (21883 total)", "last_block": "#291883 (66d03f819dde948b297c8d582889246d7ba11a5b947335497f8716a7b608f78e)"}
```

> **Note** Alternatively, you can simply run `./devel/local/start.sh` which runs this command for you.

## Running with your own substreams

1. Add a 'map' module to your `substreams.yaml` with an output type of `proto:sf.substreams.kv.v1.KVOperations`:

```yaml
modules:
  - name: kv_out
    kind: map
    inputs:
      - store: your_store
    output:
      type: proto:substreams.kv.v1.KVOperations
```

2. Write the `kv_out` module in your substreams code (and recompile)

* See [substreams-eth-block-meta](https://github.com/streamingfast/substreams-eth-block-meta) for an example integration


3. Run your substreams with the sink-kv, writing a local 'badger' database

> **Note** To connect to substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

```shell
substreams-sink-kv run \
    "badger3:///$(pwd)/my-badger.db" \
    "mainnet.eth.streamingfast.io:443" \
    "substreams.yaml" \
    kv_out
```

## Querying the data

### In command-line

* Get single block data: `grpcurl --plaintext   -d '{"key":"month:last:201511"}' localhost:8000 sf.substreams.sink.kv.v1.Kv/GetMany`
* Get many: `grpcurl --plaintext   -d '{"keys":["day:first:20151201","day:first:20151202"]}' localhost:8000 sf.substreams.sink.kv.v1.Kv/GetMany`
* By prefix: `grpcurl --plaintext   -d '{"prefix": "day:first:201511", "limit":31}' localhost:8000 sf.substreams.sink.kv.v1.Kv/GetByPrefix`
* Scan: `grpcurl --plaintext   -d '{"begin": "day:first:201501", "exclusive_end": "day:first:2016", "limit":400}' localhost:8000 sf.substreams.sink.kv.v1.Kv/Scan`

> **Note** These commands have been tested with `devel/local/start.sh` running in another terminal, which runs [substreams-eth-block-meta](https://github.com/streamingfast/substreams-eth-block-meta)
### From your browser

* Run `npm run dev` from within `/connect-web-example` to see a Vite JS example connecting to the KV GRPC Connect-Web interface

> **Note** The 'connect-web-version' works best with `devel/local/start.sh` running in another terminal, which runs [substreams-eth-block-meta](https://github.com/streamingfast/substreams-eth-block-meta)


## Protobuf generation

* Requires 'buf.build' with protoc-gen-go and protoc-gen-go-grpc (https://docs.buf.build/tour/generate-go-code)
* Run `cd proto && buf generate`

## Cargo crate

A library containing the protobuf bindings and a few helpers has been published as a crate for your convenience.

* https://crates.io/crates/substreams-sink-kv

See [substreams-eth-block-meta](https://github.com/streamingfast/substreams-eth-block-meta) for an example integration

## Contributing

For additional information, [refer to the general StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md).

## License

The `substreams-sink-kv` tool [uses the Apache 2.0 license](https://github.com/streamingfast/substreams/blob/develop/LICENSE/README.md).

This is a command line tool to quickly sync a substreams with a kv database.
