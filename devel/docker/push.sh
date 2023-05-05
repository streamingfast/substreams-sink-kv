#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

main() {
  pushd "$ROOT" &> /dev/null

  while getopts "h" opt; do
    case $opt in
      h) usage && exit 0;;
      \?) usage_error "Invalid option: -$OPTARG";;
    esac
  done
  shift $((OPTIND-1))

  docker_file="$1"; shift
  if [[ $docker_file == "" ]]; then
    docker_file="Dockerfile.goreleaser"
  fi

  docker build -f "$docker_file" . -t goreleaser-local:v1.20.2
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
  echo "usage: push.sh <Dockerfile>"
  echo ""
  echo "Builds and push the following Dockefile with the correct tags."
  echo ""
  echo "Options"
  echo "    -h          Display help about this script"
}

main "$@"