#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

main() {
  pushd "$ROOT" &> /dev/null
    pushd "blockmeta_wasm_query" &> /dev/null
      ./build.sh
    popd  >/dev/null
  popd  >/dev/null
}

main "$@"
