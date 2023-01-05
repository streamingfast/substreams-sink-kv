# substreams-sink-kv

This is a command line tool to quickly sync a substreams with a kv database.

### Quickstart

1. Install `substreams-sink-kv` (Installation from source required for now):

 ```bash
 go install ./cmd/substreams-sink-kv
 ```

2. Add a 'map' module to your `substreams.yaml` with an output type of `proto:substreams.database.v1.DatabaseChanges`:

    modules:
      - name: db_out
        kind: map
        initialBlock: 0
        inputs:
          - source: sf.ethereum.type.v2.Block
          - store: store_something
        output:
          type: proto:substreams.database.v1.DatabaseChanges

3. Run the sink to a local 'badger' database

    > To connect to substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

    ```shell
    substreams-sink-kv run \
        "badger:///home/user/sf-data/my-badger.db" \
        "mainnet.eth.streamingfast.io:443" \
        "substreams.yaml" \
        db_out
    ```
