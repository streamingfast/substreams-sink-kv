#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

main() {
  cd "$ROOT" &> /dev/null

  while getopts "hbc" opt; do
    case $opt in
      h) usage && exit 0;;
      c) clean=true;;
      \?) usage_error "Invalid option: -$OPTARG";;
    esac
  done
  shift $((OPTIND-1))

  set -e

  if [[ "$clean" == "true" ]]; then
    echo "Cleaning up existing data"
    rm badger_data
  fi

  pg_dsn="${KV_DSN:-"badger:///${ROOT}/badger_data.db"}"
  sink="../substreams-sink-kv"

  $sink run \
    ${pg_dsn} \
    "${SUBSTREAMS_ENDPOINT:-"mainnet.eth.streamingfast.io:443"}" \
    "${SUBSTREAMS_MANIFEST:-"substreams-eth-block-meta-v0.2.0.spkg"}" \
    "${SUBSTREAMS_MODULE:-"db_out"}" \
    "$@"
}

main "$@"
