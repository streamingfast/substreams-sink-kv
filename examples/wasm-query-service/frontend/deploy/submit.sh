#!/bin/bash -e

set -u
export VITE_API_URL="https://wasm-query-kv-demo.mainnet.eth.streamingfast.io/api/"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
set +e
pushd "$SCRIPT_DIR"
    rm -rf dist
    yarn build
    TAG=$(git rev-parse --short --dirty HEAD)
    docker buildx build --platform linux/amd64 -t gcr.io/eoscanada-shared-services/wasm-query-example:${TAG} --push ..
popd
