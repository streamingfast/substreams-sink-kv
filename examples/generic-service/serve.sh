#!/usr/bin/env bash

export DYLD_LIBRARY_PATH=$LIBRARY_PATH
substreams-sink-kv serve "badger3://$(pwd)/badger_data.db" substreams.yaml --listen-addr="localhost:8000"
