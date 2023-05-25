# Substreams key-value service Sink

A [Substreams _sink_](https://substreams.streamingfast.io/developers-guide/sink-targets) to pipe data from a [Substreams](https://substreams.streamingfast.io) endpoint into a key-value store and serve queries through either:
- [Connect-Web protocol](https://connect.build/docs/introduction) (gRPC-compatible) via the `GenericService`
- a User defined WASM query service

## Requirements

##### WasmEdge

Learn about WasmEdge from its [Quick Start Guide](https://wasmedge.org/book/en/quick_start/install.html), or simply run the following to install.

```bash
curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- --version 0.11.2
```

## Install

Get from the [Releases tab](https://github.com/streamingfast/substreams-sink-kv/releases), or from source:

```bash
go install -v github.com/streaminfast/substreams-sink-kv/cmd/substreams-sink-kv@latest
```

## Running

The `substreams-sink-kv` binary offers two mode of operations:
- `inject`: Runs a Substreams and pipes the data into a key-value store
- `serve`: Serves data through a query service: `GenericService` or `WASMQueryService`

You can run `substreams-sink-kv` solely in `inject` or `serve` mode or both.

> **Note** To connect to Substreams you will need an authentication token, follow this [guide](https://substreams.streamingfast.io/reference-and-specs/authentication) to obtain one.

```bash
# Inject Mode
substreams-sink-kv inject mainnet.eth.streamingfast.io:443 <kv_dsn> <substreams_spkg_path> <kv_out>

# Serve Mode
substreams-sink-kv serve <kv_dsn> <substreams_spkg_path> --listen-addr=":9000"

# Inject and Serve Mode
substreams-sink-kv inject mainnet.eth.streamingfast.io:443 <kv_dsn> <substreams_spkg_path> <kv_out> --listen-addr=":9000"
```

## Query Service

The Query Service is an API that allows you to consume data from your sinked key-value. There are 2 types of Query Services:
- Generic Service
- Wasm Query Service

the `sink` block of your Substreams manifest defines and configures which one to use

```yaml
specVersion: v0.1.0
package:
  name: "your_substreams_to_sink"
  version: v0.0.1

...

sink:
  module: kv_out
  type: sf.substreams.sink.kv.v1.WASMQueryService
  config:
    wasmQueryModule: "@@./blockmeta_wasm_query/blockmeta_wasm_query.wasm"
    grpcService: "eth.service.v1.Blockmeta"
```

breaking down the `sink` block we get the following:

- **module**: The name of the module that will be used to sink the key-value store. The module should be of kind `map` with an output type of [`sf.substreams.sink.kv.v1.KVOperations`](https://github.com/streamingfast/substreams-sink-kv/blob/main/proto/sf/substreams/sink/kv/v1/kv.proto)
- **type**: Support to types currently:
  - [`sf.substreams.sink.kv.v1.WASMQueryService`](./proto/sf/substreams/sink/kv/v1/services.proto)
  - [`sf.substreams.sink.kv.v1.GenericService`](./proto/sf/substreams/sink/kv/v1/services.proto)
- **config**: a key-value structure that matches the attributes of the Proto object for the given `type` selected above

> **_NOTE:_**  the `@@` notation will read the path and inject the content of the file in bytes, while the `@` notation will dump the file content in ascii


### Generic Service

The Generic Query service is a [Connect-Web protocol](https://connect.build/docs/introduction) (gRPC-compatible). It exposes a browser and gRPC-compatible APIs. The API is defined in `protobuf` [here](./proto/sf/substreams/sink/kv/v1/read.proto).

You can find a detailed example with documentation [here](./examples/generic-service)

### WASM Query Service

The wasm query service is a user-defined gRPC API that is backed by WASM code, which has access to underlying key-value store.

You can find a detailed example with documentation [here](./examples/wasm-query-service)

## Contributing

Refer to the [general StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md).

### WasmEdge

Bumping `github.com/second-state/WasmEdge-go` to a newer version requires matching WasmEdge version. For example `github.com/second-state/WasmEdge-go@0.11.2` requires [WasmEdge 0.11.2](https://github.com/WasmEdge/WasmEdge/releases/tag/0.11.2).

If a new version of [WasmEdge](https://github.com/WasmEdge/WasmEdge), it requires also modifying the [./devel/docker/Dockerfile.goreleaser](./devel/docker/Dockerfile.goreleaser) file so it pulls dependencies required for building using the correct version. Testing the build can be done using by first building the Docker image via `./devel/docker/push.sh` and then running:

```bash
docker run --rm -it -v `pwd`:/work -w /work goreleaser-wasmedge:v1.20.3 release --snapshot --clean --skip-validate
```

Manual Golang compilation within the docker image can be done with:

```bash
docker run --rm -it -v `pwd`:/work -w /work --entrypoint bash goreleaser-wasmedge:v1.20.3
cd /work
# Adjust CC, CXX, C_INCLUDE_PATH, LIBRARY_PATH and GOOS/GOARCH according to .goreleaser.yaml file
CGO_ENABLED=1 CC=oa64-clang CXX=oa64-clang++ GOOS=darwin GOARCH=arm64 C_INCLUDE_PATH=/usr/local/osxcross/include/arm64 LIBRARY_PATH="/usr/local/osxcross/lib/arm64" go build -trimpath -mod=readonly -ldflags="-s -w" -o /work/substreams-sink-kv-cross-compiled ./cmd/substreams-sink-kv/
```

## `kvdb` tool

You can inspect badger files written by this tool with:

```
go install github.com/streamingfast/kvdb/cmd/kvdb@develop
```


## License

[Apache 2.0](https://github.com/streamingfast/substreams/blob/develop/LICENSE/README.md).

