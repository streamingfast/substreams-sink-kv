#!/bin/bash
# Copyright 2021 dfuse Platform Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd ../.. && pwd )"

# Protobuf definitions
PROTO=${2:-"$ROOT/proto"}
PROTO_OUT="$ROOT/src/pb"

GEN_FILE="$PROTO_OUT/last_generate.txt"

function main() {
  checks

  set -e

  cd "$ROOT" &> /dev/null

  echo "Generating proto files"
  generate "substreams/sink/kv/v1/kv.proto"
  generate "substreams/sink/kv/v1/types.proto"
  echo "Successfully generated proto files"

  echo "generate.sh - $(date) - $(whoami)" > "$GEN_FILE"
}

# usage:
# - generate <protoPath>
# - generate <protoBasePath/> [<file.proto> ...]
function generate() {
    base=""
    if [[ "$#" -gt 1 ]]; then
      base="$1"; shift
    fi

    for file in "$@"; do
      protoc -I"$PROTO" "$base""$file" --prost_out="$PROTO_OUT"
    done
}

function checks() {
  result=$(printf "" | protoc --version 2>&1 | grep -Eo "[0-9]\.\w+")
  if [[ "$result" == "" ]]; then
    echo "Your version of 'protoc' (at $(which protoc)) is not recent enough."
    echo ""
    echo "To fix your problem, visit official protoc installation site: https://grpc.io/docs/protoc-installation/"
    echo ""
    echo "If everything is working as expected, the command:"
    echo ""
    echo "  protoc --version"
    echo ""
    echo "Should print a version over 3 'libprotoc 3.19.4' (if it just hangs, you don't have the correct version)"
    exit 1
  fi
}

main "$@"
