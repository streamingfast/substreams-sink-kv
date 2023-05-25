# Generic Service example

In this example, we will launch the [`block-meta` substream](https://github.com/streamingfast/substreams-eth-block-meta), sink it to a key-value store and launch a Generic service gRPC API and connect it to a webpage.

## Requirements

##### WasmEdge

Learn about WasmEdge from its [Quick Start Guide](https://wasmedge.org/book/en/quick_start/install.html), or simply run the following to install.

```bash
curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- --version 0.11.2
```

> **Note** If you use `zsh`, the final installation instructions talks about sourcing `"$HOME/.zprofile` but it seems this file is not created properly in all cases. If it's the case, add `source "$HOME/.wasmedge/env"` at the end of your `.zshrc` file.

##### Buf CLI

`buf` command v1.11.10 or later https://docs.buf.build/installation

```bash
brew install bufbuild/buf/buf
```

##### `npm` and `nodejs`

See https://nodejs.org for installation instructions.

## Install

Get from the [Releases tab](https://github.com/streamingfast/substreams-sink-kv/releases), or from source:

```bash
go install -v github.com/streaminfast/substreams-sink-kv/cmd/substreams-sink-kv
```

### Substreams

The `block-meta` Substreams tracks the first and last block of every month since genesis block. The Substreams has a `map` module with an output type of `sf.substreams.sink.kv.v1.KVOperations`

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

The module outputs  [`KVOperations`](../../proto/sf/substreams/sink/kv/v1/kv.proto) that the `substreams-sink-kv` will apply to key-value-store. The implementation details can be found [here](https://github.com/streamingfast/substreams-eth-block-meta/blob/adfd451a8354eba1fa40e94dc205b1499df69f5b/src/kv_out.rs)

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

> **Note** To connect to Substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one,

You can run the `substreams-sink-kv` inject mode.

```bash
# Required only on MacOS to properly instruct the 'substreams-sink-kv' where to find the WasmEdge library
export DYLD_LIBRARY_PATH=$LIBRARY_PATH

substreams-sink-kv inject mainnet.eth.streamingfast.io:443 "badger3://$(pwd)/badger_data.db" substreams.yaml
```
> **Note** You can also use the `inject.sh` scripts which contains the call above

The `inject` mode is running the [`block-meta` substream](https://github.com/streamingfast/substreams-eth-block-meta) and applying the `KVOperation` to a local `badger -b` that is located here `./badger_data.db`.

After a few minutes of sinking your local `badger-db` should contains keys. You can close the `inject` process.

We can introspect the store with our [`kvdb` CLI](https://github.com/streamingfast/kvdb)

```bash
  kvdb read prefix kmonth:first --dsn "badger3://$(pwd)/badger_data.db" --decoder="proto://./proto/block_meta.proto@eth.block_meta.v1.BlockMeta"
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

### Generic Service

The Generic Query service is a [Connect-Web protocol](https://connect.build/docs/introduction) (gRPC-compatible). It exposes a browser and gRPC-compatible APIs.

The API is defined in `protobuf` [here](../../proto/sf/substreams/sink/kv/v1/read.proto).


Launch the API, this starts the Connect-Web server

```bash
  export DYLD_LIBRARY_PATH=$LIBRARY_PATH
  substreams-sink-kv serve "badger3://$(pwd)/badger_data.db" substreams.yaml --listen-addr=":8080"
```

> **Note** You can also use the `serve.sh` scripts which contains the call above

### Website

We create a straightforward web app template:

```bash
npm create vite@latest -- connect-web-example --template react-ts
cd connect-web-example/
npm install
npm install --save-dev @bufbuild/protoc-gen-connect-web @bufbuild/protoc-gen-es
npm install @bufbuild/connect-web @bufbuild/protobuf
```

Create the  `buf.gen.yaml`, that will configure `buf` to generate our `typescript` files based on the `proto` definitions

```yaml
version: v1
plugins:
  - plugin: es
  - plugin: connect-web
```

* Add script line to generate `typescript` files from the `proto` files

```json
    # package.json
    # "scripts": {
    # ...
    "buf:generate": "buf generate ../proto/sf/substreams/sink/kv/v1 && buf generate ./proto",
```

* Generate code:

`npm run buf:generate` to generate the following files under `gen/` : `kv_pb.ts`, `read_connectweb.ts`, `read_pb.ts`

* Create client from KV in App.tsx:

```ts
import {
    createConnectTransport,
    createPromiseClient,
} from "@bufbuild/connect-web";
import { Kv } from "../gen/read_connectweb";
const transport = createConnectTransport({
    baseUrl: "http://localhost:8000",
});
const client = createPromiseClient(Kv, transport);

```

* Call client from button action:

```ts
const response = await client.get({
    key: inputValue,
});

// then display `response.value`, typed as 'Uint8Array'
```

* Use our generated 'block_meta_pb' protobuf bindings to decode the value:

```ts
import { BlockMeta } from "../gen/block_meta_pb";

const blkmeta = BlockMeta.fromBinary(response.value);
output = JSON.stringify(blkmeta, null, 2);
```

The rest is just formatting, error-handling and front-end stuff ...

Lastly, we need to run our frontend in another terminal from our API

```bash
npm run dev
```

## Reference

- Connect-web code generation [https://connect.build/docs/web/generating-code](https://connect.build/docs/web/generating-code)
