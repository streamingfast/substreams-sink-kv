#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export DYLD_LIBRARY_PATH=$LIBRARY_PATH
substreams-sink-kv serve "badger3://$(pwd)/badger_data.db" "$ROOT/substreams.yaml"
