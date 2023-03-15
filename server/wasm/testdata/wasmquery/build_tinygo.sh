#!/bin/bash -xe

# See https://wasmedge.org/book/en/sdk/go/function.html
# And https://wasmedge.org/book/en/write_wasm/rust/bindgen.html

/Users/abourget/dev/tinygo/build/tinygo build -scheduler=none -o hello_wasm_query.wasm -target wasi hello_wasm_query.go
wasm2wat ./hello_wasm_query.wasm > hello_wasm_query_tinygo.wat
