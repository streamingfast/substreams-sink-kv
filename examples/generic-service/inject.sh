#!/usr/bin/env bash

export DYLD_LIBRARY_PATH=$LIBRARY_PATH
substreams-sink-kv inject -e mainnet.eth.streamingfast.io:443 "badger3://$(pwd)/badger_data.db" substreams.yaml
