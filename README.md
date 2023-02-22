# Substreams key-value service Sink

A [Substreams _sink_](https://substreams.streamingfast.io/developers-guide/sink-targets) to pipe data from a [Substreams](https://substreams.streamingfast.io) endpoint into a key-value store and serve queries through [Connect-Web protocol](https://connect.build/docs/introduction) (gRPC-compatible).

## Install

Get from the [Releases tab](https://github.com/streamingfast/substreams-sink-kv/releases), or from source:

```bash
go install -v github.com/streaminfast/substreams-sink-kv/cmd/substreams-sink-kv
```

## Run

> **Note** To connect to substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

```bash
export PACKAGE=https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg

substreams-sink-kv run "badger3://$(pwd)/badger_data.db" mainnet.eth.streamingfast.io:443 $PACKAGE kv_out
```

## Integrate

Create a [Substreams `map` module](https://substreams.streamingfast.io) with an output type of [`sf.substreams.sink.kv.v1.KVOperations`](https://github.com/streamingfast/substreams-sink-kv/blob/main/proto/substreams/sink/kv/v1/kv.proto):

```yaml
modules:
  - name: kv_out
    kind: map
    inputs:
      - store: your_store
    output:
      type: proto:substreams.kv.v1.KVOperations
```

**Cargo.toml** (see [crate](https://crates.io/crates/substreams-sink-kv))

```toml
[dependencies]
...
substreams-sink-kv = "0.1.1"
```

**`lib.rs`**

```rust
use substreams::proto;
use substreams::store::{self, DeltaProto};
use substreams_sink_kv::pb::kv::KvOperations;

use crate::pb::block_meta::BlockMeta;

pub fn block_meta_to_kv_ops(ops: &mut KvOperations, deltas: store::Deltas<DeltaProto<BlockMeta>>) {
    use substreams::pb::substreams::store_delta::Operation;

    for delta in deltas.deltas {
        match delta.operation {
            Operation::Create | Operation::Update => {
                let val = proto::encode(&delta.new_value).unwrap();
                ops.push_new(delta.key, val, delta.ordinal);
            }
            Operation::Delete => ops.push_delete(&delta.key, delta.ordinal),
            x => panic!("unsupported opeation {:?}", x),
        }
    }
}
```

More details here: https://github.com/streamingfast/substreams-eth-block-meta/blob/master/src/kv_out.rs


## Query

```bash
grpcurl --plaintext -d '{"key":"month:last:201511"}' localhost:8000 \ 
    sf.substreams.sink.kv.v1.Kv/Get
grpcurl --plaintext -d '{"keys":["day:first:20151201","day:first:20151202"]}' localhost:8000 \ 
    sf.substreams.sink.kv.v1.Kv/GetMany
grpcurl --plaintext -d '{"prefix": "day:first:201511", "limit":31}' localhost:8000 \
    sf.substreams.sink.kv.v1.Kv/GetByPrefix
grpcurl --plaintext -d '{"begin": "day:first:201501", "exclusive_end": "day:first:2016", "limit":400}' localhost:8000 \
    sf.substreams.sink.kv.v1.Kv/Scan
```

Sample browser interface:

```bash
# in a first tab
./devel/local/start.sh

# in a second tab
cd connect-web-example
npm run dev
```


## Contributing

Refer to the [general StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md).

## License

[Apache 2.0](https://github.com/streamingfast/substreams/blob/develop/LICENSE/README.md).

