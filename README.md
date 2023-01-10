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

### Protobuf generation

* Requires 'buf.build' with protoc-gen-go and protoc-gen-go-grpc (https://docs.buf.build/tour/generate-go-code)
* Run `cd proto && buf generate`

## Cargo crate

Note: this crate is a stub, mostly to contain the protobuf bindings, feel free to use it or not.

* https://crates.io/crates/substreams-sink-kv

### Example integration

* See https://github.com/streamingfast/substreams-eth-block-meta
