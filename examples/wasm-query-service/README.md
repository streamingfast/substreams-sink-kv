# Wasm Query Service example

In this example we will launch the [`block-meta` substream](https://github.com/streamingfast/substreams-eth-block-meta), sink it to a key-value store and launch a WASM Query.


### Substream
The `block-meta` substreams tracks the first and last block of every month since genesis block. The substreams has a `map` module with an output type of `sf.substreams.sink.kv.v1.KVOperations`

```yaml
...
- name: kv_out
  kind: map
  inputs:
    - store: store_block_meta_start
      mode: deltas
    - store: store_block_meta_end
      mode: deltas
  output:
    type: proto:sf.substreams.sink.kv.v1.KVOperations
...
```
You can see the full `substreams.yaml` [here](https://github.com/streamingfast/substreams-eth-block-meta/blob/adfd451a8354eba1fa40e94dc205b1499df69f5b/substreams.yaml#L46-L54)

The output module emits [`KVOperations`](../../proto/substreams/sink/kv/v1/kv.proto) that the `substreams-sink-kv` will apply to key-value-store. The implementation details can be found [here](https://github.com/streamingfast/substreams-eth-block-meta/blob/adfd451a8354eba1fa40e94dc205b1499df69f5b/src/kv_out.rs)

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

### Running Injector

> **Note** To connect to substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

You can run the `substreams-sink-kv` inject mode with the `start-inject.sh` script. The script is essentially running this command:

```bash
 substreams-sink-kv inject -e mainnet.eth.streamingfast.io:443 "badger3://$(pwd)/badger_data.db" substreams.yaml
```

The `substreamd-sink` is running the `block-meta` substreams and applying the `KVOperation` to a local `badger db` that is `./badger_data.db`. 

Close the process after a few minutes of sinking at which your local badger key-value store should contains keys. We can introspec the store with our [`kvdb` CLI](https://github.com/streamingfast/kvdb)

```bash
  kvdb read prefix kmonth:first --dsn "badger3://$(pwd)/badger_data.db" --decoder="proto://./blockmeta_wasm_query/proto/block_meta.proto@eth.block_meta.v1.BlockMeta"
```

You should get an output like this

```bash
keys with prefix: kmonth:first
kmonth:first:197001	->	{"hash":"1OVnQPh2rvjAELhqQNX1Z0WhGNCQajTmmuyMDbHLj6M=","parentHash":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","timestamp":"1970-01-01T00:00:00Z"}
kmonth:first:201507	->	{"number":"1","hash":"iOltRTe+pNnAXRJUmQezJWHTvzH0Wq5zTNwRnxNAbLY=","parentHash":"1OVnQPh2rvjAELhqQNX1Z0WhGNCQajTmmuyMDbHLj6M=","timestamp":"2015-07-30T15:26:28Z"}
kmonth:first:201508	->	{"number":"13775","hash":"Lc7K1M8gedGBacoFvCHnugrdcTK5OCmEdg9D8nYb2CI=","parentHash":"q6q7j4t/f6B2aPs4/VoI2pgUzYrRink+VO72+pt5SrQ=","timestamp":"2015-08-01T00:00:03Z"}
kmonth:first:201509	->	{"number":"170395","hash":"BcQ5Q4WDvwWinOLmvmBRow+93ncq4UPzlyaoFzETBas=","parentHash":"PrACXrHJMjr8l/zN8E4mucPsxTEVMnAj9uog69E75jo=","timestamp":"2015-09-01T00:00:20Z"}
kmonth:first:201510	->	{"number":"314573","hash":"Po8RfewImJMcUEK0KDdDzORKt3mqx2VihhfA3EyRvAI=","parentHash":"Sek9sYxwxk2MuiJI4V8H/D6IX/unSxICfVYVBP9hpx0=","timestamp":"2015-10-01T00:00:17Z"}
kmonth:first:201511	->	{"number":"470668","hash":"1AB3cXZtMx5JL2KLz4dCBpjq+lsZEf6zgHd14+u+z1k=","parentHash":"6Kq+SBHeXyRdZ+/rC6he5QZWKHVPCoo0ebbgyUuL6H4=","timestamp":"2015-11-01T00:00:08Z"}
kmonth:first:201512	->	{"number":"622214","hash":"fw3ZOpMrUo8mqZReGkt+SBfnpv0aiPkKF2qrdmZn27o=","parentHash":"cPTq4v4Q7Ys5ivJjdiaxEjES4SIKRkZV238e3LhbQFU=","timestamp":"2015-12-01T00:00:01Z"}
...
````

### WASM Query

The wasm query service, is a user defined service that exposes a consumable gRPC API. There are 2 importants parts:

- The `.proto` that finds the gRPC api. In this example you can find the proto file [here](./blockmeta_wasm_query/proto/service.proto)
```protobuf
...
service Blockmeta {
  rpc GetMonth(GetMonthRequest) returns (stream MonthResponse);
//  rpc GetYear(GetYearRequest) returns (stream YearReponse);
}
...
```

- The `WASM` code that is executed when the gRPC API is called. The full implementation is [here](./blockmeta_wasm_query/src/lib.rs)

The important things to note are:

1) For every `method` in the defined gRPC service (i.e. `GetMonth`) there needs to be a matching `WASM` query function where the name is the fully qualified gRPC method name sanitized (all lower case and period and replaced by underscore). For example `eth.service.v1.blockmeta.GetMonth` execute this WASM function `eth_service_v1_blockmeta_getmonth`
2) the WASM function has access to the underlying key-value store and can perform common key-value operation `get`, `getMany`, `prefix`, `scan`

Launching the `substreams-sink-kv serve` command will essentially start a gRPC service that exposes the defined `.proto` service backed by the `WASM` code you wrote.

```bash
  substreams-sink-kv serve "badger3://$(pwd)/badger_data.db" substreams.yaml
```

In a separate terminal you can run the following command, to consume your gRPC API

```bash
  grpcurl -plaintext -proto ./blockmeta_wasm_query/proto/service.proto -d '{"year": "2019","month":"05"}' localhost:7878 eth.service.v1.Blockmeta.GetMonth
```


