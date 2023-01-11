# Web example

* Reference:  https://connect.build/docs/web/generating-code

# Requirements:

* buf.build's `buf` command v1.11.0 or later
* `npm` and `nodejs` with decent version

# Steps taken to create this example

* Straightforward web app template creation:
```
npm create vite@latest -- connect-web-example --template react-ts
cd connect-web-example/
npm install
npm install --save-dev @bufbuild/protoc-gen-connect-web @bufbuild/protoc-gen-es
npm install @bufbuild/connect-web @bufbuild/protobuf
```

* Create `buf.gen.yaml`  (note: this requires a buf version 1.11.0 or later)

```
version: v1
plugins:
  - plugin: es
  - plugin: connect-web
```

* Add script line to generate code:

```
    # package.json
    # "scripts": {
    # ...
    "buf:generate": "buf generate ../proto/substreams/sink/kv/v1",
```

* Generate code:

`npm run buf:generate` to generate the following files under `gen/` : `kv_pb.ts`, `read_connectweb.ts`, `read_pb.ts`

* Create client from KV in App.tsx:

```
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
```
const response = await client.get({
    key: inputValue,
});

// then display `response.value`, typed as 'Uint8Array'
```

The rest is just front-end formatting ...

# Run it

* `npm run dev`
