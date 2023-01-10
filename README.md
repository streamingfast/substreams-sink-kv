# substreams-sink-kv

This is a command line tool to quickly sync a substreams with a kv database.

### Quickstart

1. Install `substreams-sink-kv` (Installation from source required for now):

 ```bash
 go install ./cmd/substreams-sink-kv
 ```

2. Add a 'map' module to your `substreams.yaml` with an output type of `proto:substreams.kv.v1.KVOperations`:

    modules:
      - name: kv_out
        kind: map
        initialBlock: 0
        inputs:
          - store: store_something
        output:
          type: proto:substreams.kv.v1.KVOperations

3. Run the sink to a local 'badger' database

    > To connect to substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

    ```shell
    substreams-sink-kv run \
        "badger3:///home/user/sf-data/my-badger.db" \
        "mainnet.eth.streamingfast.io:443" \
        "substreams.yaml" \
        kv_out
    ```

### Even quicker iteration

* Run `devel/local/start.sh`

### Querying the data

#### In command-line

(these commands have been tested while running the 'devel/local/start.sh' in another terminal)

* Get single block data: `grpcurl --plaintext   -d '{"key":"72e08a117f596b31ae49bda6f093c2d812e64f69bb3d5aab2f4f47fdaea0c512"}' localhost:8000 substreams.sink.kv.v1.Kv/GetMany`
* Get many: `grpcurl --plaintext   -d '{"keys":["72e08a117f596b31ae49bda6f093c2d812e64f69bb3d5aab2f4f47fdaea0c512","72e046b07eb0263277e3a5e5968d60810273f74a1e8ef64c9b3bf522285accc8"]}' localhost:8000 substreams.sink.kv.v1.Kv/GetMany`
* By prefix: `grpcurl --plaintext   -d '{"prefix": "ffa", "limit":100}' localhost:8000 substreams.sink.kv.v1.Kv/GetByPrefix`
* Scan: `grpcurl --plaintext   -d '{"begin": "ffff46b07eb0263277e3a5e5968d60810273f74a1e8ef64c9b3bf522285accc8", "exclusive_end": "ffffb5738d3f365313963d49af78f08e1953623189bea23d5ed3562a55f8d165", "limit":3}' localhost:8000 substreams.sink.kv.v1.Kv/Scan`

#### From your browser

(this has been tested while running the 'devel/local/start.sh' in a terminal)

* Run `npm run dev` from within `/connect-web-example` to see a Vite JS example connecting to the KV GRPC Connect-Web interface

### Protobuf generation

* Requires 'buf.build' with protoc-gen-go and protoc-gen-go-grpc (https://docs.buf.build/tour/generate-go-code)
* Run `cd proto && buf generate`

## Cargo crate

Note: this crate is a stub, mostly to contain the protobuf bindings, feel free to use it or not.

* https://crates.io/crates/substreams-sink-kv

### Example integration

* See https://github.com/streamingfast/substreams-eth-block-meta
