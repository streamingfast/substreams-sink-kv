#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
sink="$ROOT/../substreams-sink-kv"
main() {
  cd "$ROOT" &> /dev/null

  while getopts "hbc" opt; do
    case $opt in
      h) usage && exit 0;;
      c) clean=true;;
      b) force_build=true;;
      \?) usage_error "Invalid option: -$OPTARG";;
    esac
  done
  shift $((OPTIND-1))

  set -e

  wasmfile="./blockmeta_wasm_query/blockmeta_wasm_query.wasm"
  if [[ ! -f "$wasmfile" ]]; then
    echo "$wasmfile does not exists, will force a build"
    force_build=true;
  fi

  if [[ "$force_build" == "true" ]]; then
    echo "Building WASM query"
    pushd "$ROOT/blockmeta_wasm_query" &> /dev/null
      ./build.sh
    popd  >/dev/null
    echo "WASM query build complete"
    echo ""
  fi

  dsn="${KV_DSN:-"badger3:///${ROOT}/badger_data.db"}"

  $sink serve \
    ${dsn} \
    "substreams.yaml" \
    "$@"
}

main "$@"
