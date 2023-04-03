#!/bin/bash -e

NETWORK=$1
set -u
export VITE_API_URL=$NETWORK

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
set +e
pushd "$SCRIPT_DIR"
    rm -rf dist
    yarn build
    TAG=$(git rev-parse --short --dirty HEAD)
    docker build -t gcr.io/eoscanada-shared-services/wasm-query-example:${TAG} ..
    docker push gcr.io/eoscanada-shared-services/wasm-query-example:${TAG}
popd

#deployment=website-lostworld-staging
#if [ "$NETWORK" == "mainnet" ]; then
#    deployment=website-lostworld
#fi
#
#CMD="kubectl -n lostworld set image deploy/$deployment nginx=gcr.io/eoscanada-shared-services/lostworld:$NETWORK-${TAG}"
#echo "Running: $CMD"
#eval $CMD