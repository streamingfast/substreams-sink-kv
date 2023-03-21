#!/bin/bash -e

cargo build --target wasm32-wasi --release
cp target/wasm32-wasi/release/wasm_query.wasm ./blockmeta_wasm_query.wasm
