#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

tag_host="ghcr.io/streamingfast/"

main() {
  pushd "$ROOT" &> /dev/null

  while getopts "hl" opt; do
    case $opt in
      h) usage && exit 0;;
      l) tag_host="";;
      \?) usage_error "Invalid option: -$OPTARG";;
    esac
  done
  shift $((OPTIND-1))

  docker_file="$1"; shift
  if [[ $# -ge 1 ]]; then
    version="$1"; shift
  fi

  if [[ ! -f "$docker_file" ]]; then
    usage_error "Provided Dockerfile '$docker_file' does not exist"
  fi

  if [[ "$version" == "" ]]; then
    version=`git -c 'versionsort.suffix=-' ls-remote --exit-code --refs --sort='version:refname' --tags origin '*.*.*' | tail -n1 | tr "\t" " " | cut -d ' ' -f 2 | sed 's|refs/tags/||g'`
  fi

  name="`printf $docker_file | sed 's/Dockerfile.//g'`"

  if [[ "$name" == "runtime" ]]; then
    build_args="--build-arg='VERSION=$version' --platform=linux/amd64 -t '${tag_host}substreams-sink-kv:$version'"
  elif [[ "$name" == "builder" ]]; then
    build_args="--build-arg='VERSION=$version' --platform=linux/amd64,linux/arm64 -t ${tag_host}substreams-sink-kv-builder:v1.20.3"
  else
    usage_error "Unknown Dockerfile '$docker_file'"
  fi

  docker buildx build -f "$docker_file" . --push $build_args
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
  echo "usage: push.sh <dockerfile> [<version>]"
  echo ""
  echo "Builds and push the following Dockefile with the correct tags."
  echo ""
  echo "Options"
  echo "    -h          Display help about this script"
}

main "$@"