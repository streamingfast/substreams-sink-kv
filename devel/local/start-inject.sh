#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
sink="$ROOT/../substreams-sink-kv"
echo $ROOT
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

  if [[ "$clean" == "true" ]]; then
    echo "Cleaning up existing data"
    rm -rf badger_data.db
    echo ""
  fi

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

  $sink inject \
    -e "${SUBSTREAMS_ENDPOINT:-"mainnet.eth.streamingfast.io:443"}" \
    ${dsn} \
    "substreams.yaml" \
    "$@"
}

usage_error() {
  message="$1"
  exit_code="$2"

  echo "ERROR: $message"
  echo ""
  usage
  exit ${exit_code:-1}
}

usage() {
  echo "usage: start-inject.sh <all>"
  echo ""
  echo "Runs a substreams-sink-kv in inject mode"
  echo ""
  echo "Options"
  echo "    -c          Cleans key-value store"
  echo "    -b          Force builds the wasm query"
  echo "    -h          Display help about this script"

}

main "$@"
