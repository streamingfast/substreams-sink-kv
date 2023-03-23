# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3](https://github.com/streamingfast/substreams-sink-kv/releases/tag/v0.1.3)

### Added 

* Added WASM Query Support
* `GenericService` example `./examples/generic-service`
* `WASMQueryService` example `./examples/wasm-query-service`
* Panic handling in wasm query engine 

### Changed

* Removed support for GRPC `stream`, only support request response
* Remove `substream-sink-kv run` CLI and added `substream-sink-kv inject` & `substream-sink-kv serve`
* Move `rust` crate to its own repo `https://github.com/streamingfast/substreams-sink-rs`

## [0.1.2](https://github.com/streamingfast/substreams-sink-kv/releases/tag/v0.1.2)

### Changed

* Prevent hanging on connection error.

### Fixed

* Fixed wrong accepted module's output type `sf.substreams.kv.v1.KVOperations`, correct value is `sf.substreams.sink.kv.v1.KVOperations`.

## [0.1.1](https://github.com/streamingfast/substreams-sink-kv/releases/tag/v0.1.1)

* Initial release