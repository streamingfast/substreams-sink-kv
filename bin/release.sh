#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

# Provide by the user
publish=""

# Constants
golang_cross_version="v1.19.4"
package_name="github.com/streamingfast/substreams-sink-kv"
substreams_name="substreams-sink-kv"

main() {
  pushd "$ROOT" &> /dev/null

  while getopts "hpw" opt; do
    case $opt in
      h) usage && exit 0;;
      p) publish="true";;
      w) review_web="true";;
      \?) usage_error "Invalid option: -$OPTARG";;
    esac
  done
  shift $((OPTIND-1))

  verify_github_token
  verify_gh
  verify_skip

  # We do not sign releases anymore because they are done in a Docker env now
  # so some adaptation is required, probably GPG support
  #verify_keybase

  version="$1"; shift

  # 'shift' can fail above, so we put bash in non failure mode after calling it
  set -e

  while true; do
   if [[ "$version" == "" ]]; then
      printf "What version do you want to release (current latest is `git describe --tags --abbrev=0`)? "
      read version
    fi

    if [[ ! "$version" =~ ^v ]]; then
      echo "Version $version is invalid, must start with a 'v'"
    else
      break
    fi
  done

  echo "Pushing to ensure GitHub knowns about the latest commit(s)"
  git push
  echo ""

  mode="Dry run, builds artifact and creates a GitHub release in draft mode, use -p flag to publish release right now"
  if [[ "$publish" == "true" ]]; then
    mode="Publishing"
  fi

  echo "About to release version $version ($mode)"
  sleep 3

  # We tag the version because goreleaser needs it to perform its work properly,
  # but we delete it at the end of the script because we will let GitHub create
  # the tag when the release is performed
  git tag "$version"
  trap "git tag -d $version > /dev/null" EXIT

  ## Substreams .spkg building
  substreams pack -o "build/${substreams_name}-${version}.spkg"
  echo ""

  ## Release Notes Generation
  start_at=$(grep -n -m 1 -E '^## .+' CHANGELOG.md | cut -f 1 -d :)
  changelod_trimmed=$(skip $start_at CHANGELOG.md | skip 1)

  # It's important to work on trimmed content to determine end because `head -n$end_at`
  # below is applied to trimmed content and thus line number found must be relative to it
  end_at=$(printf "$changelod_trimmed" | grep -n -m 1 -E '^## .+' | cut -f 1 -d :)
  printf "$changelod_trimmed" | head -n$end_at | skip 2 > .release_notes.md

  args="--rm-dist --release-notes=.release_notes.md"

	docker run \
		--rm \
		-e CGO_ENABLED=1 \
    --env-file .env.release \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v "`pwd`:/go/src/${package_name}" \
		-w "/go/src/${package_name}" \
		"goreleaser/goreleaser-cross:${golang_cross_version}" \
		$args

  echo "Release draft has been created succesuflly, but it's not published"
  echo "yet. You must now review the release and publish it if everything is"
  echo "correct."
  echo ""
  echo "Showing you the release in the terminal..."
  sleep 0.5

  args=""
  if [[ "$review_web" == "true" ]]; then
    args="--web"
  fi

  gh release view "$version" $args

  if [[ "$force" == "true" ]]; then
      publish_now="yes"
  else
    echo ""
    printf "Would you like to publish it right now? "
    read publish_now
    echo ""
  fi

  if [[ "$publish_now" == "Y" || "$publish_now" == "y" || "$publish_now" == "Yes" || "$publish_now" == "yes" ]]; then
    gh release edit "$version" --draft=false
    echo ""

    echo "Release published at $(gh release view $version --json url -q '.url')"
  else
    echo "If something is wrong, you can delete the release from GitHub"
    echo "and try again, delete the release by doing the command:"
    echo ""
    echo "  gh release delete $version"
    echo ""
    echo "You can also publish it from the GitHub UI directly, on the release"
    echo "page, press the small pencil button in the right corner to edit the release"
    echo "and then press the 'Publish release' green button (scroll down to the bottom"
    echo "of the page."
    echo ""

    printf "Do you want to open it right now? "
    read answer
    echo ""

    if [[ "$answer" == "Y" || "$answer" == "y" || "$answer" == "Yes" || "$answer" == "yes" ]]; then
      gh release view "$version" --web
    fi
  fi
}

verify_github_token() {
  release_env_file="$ROOT/.env.release"

  if [[ ! -f "$release_env_file" || ! grep -q "GITHUB_TOKEN=" "$release_env_file" ]]; then
    if [[ "$GITHUB_TOKEN" != "" ]]; then
        echo 'GITHUB_TOKEN=${GITHUB_TOKEN}' >> "$release_env_file"

    elif [[ -f "$HOME/.config/goreleaser/github_token" ]]; then

      token=$(cat "$HOME/.config/goreleaser/github_token")
      echo 'GITHUB_TOKEN=${token}' >> "$release_env_file"

    else
      echo "A '.env.release' file must be found at the root of the project and it must contain"
      echo "definition of 'GITHUB_TOKEN' variable. You need to create this file locally and the"
      echo "content should be:"
      echo ""
      echo "GITHUB_TOKEN=<your_github_token>"
      echo ""
      echo "You will need to create your own GitHub Token on GitHub website and make it available through"
      echo "the file mentioned above."

      if [[ -f "$release_env_file" ]]; then
        echo ""
        echo "Actual content of '$release_env_file' is:"
        echo ""
        cat "$release_env_file"
      fi
      exit 1
    fi
  fi
}

verify_gh() {
  if ! command -v gh &> /dev/null; then
    echo "The GitHub CLI utility (https://cli.github.com/) is required to obtain"
    echo "information about the current draft release."
    echo ""
    echo "Install via brew with 'brew install gh' or refer https://github.com/cli/cli#installation"
    echo "otherwise."
    echo ""
    echo "Don't forget to activate link with GitHub by doing 'gh auth login'."
    echo ""
    exit 1
  fi
}

verify_skip() {
  if ! command -v skip &> /dev/null; then
    echo "The 'skip' utility is required to generate the release notes from the"
    echo "changelog file."
    echo ""
    echo "Install from source using 'go install github.com/streamingfast/tooling/cmd/skip@latest'"
    echo ""
    exit 1
  fi
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
  echo "usage: release.sh [-h] [-p] [-w] [<version>]"
  echo ""
  echo "Perform the necessary commands to perform a release of the project, builds"
  echo "artifacts through Docker and publish a draft GitHub release by default."
  echo "You can use '-p' to automatically publish the release instead of in draft mode."
  echo ""
  echo "The <version> is optional, if not provided, you'll be asked the question."
  echo ""
  echo "The release being performed against GitHub, you need a valid GitHub API token"
  echo "with the necessary rights to upload release and push to repositories. It needs to"
  echo "be provided in file named '.env.release' or through an environment"
  echo "variable GITHUB_TOKEN (in which case the file is automtically created from it)."
  echo ""
  echo "The GitHub CLI utility (https://cli.github.com/) is required to obtain"
  echo "information about the current draft release. Install via brew with "
  echo "'brew install gh' or refer https://github.com/cli/cli#installation"
  echo "otherwise."
  echo ""
  echo "StreamingFast 'skip' utility is required, you can use the command"
  echo "'go install github.com/streamingfast/tooling/cmd/skip@latest' to install it"
  echo ""
  echo "You will need to have it available ('brew install keybase' on Mac OS X) and"
  echo "configure it, just setting your Git username and a password should be enough."
  echo ""
  echo "Options"
  echo "    -p          Forcing the GitHub release to be published right away instead of leaving it in draft mode"
  echo "    -w          Review the draft release within the browser instead of through the CLI"
  echo "    -h          Display help about this script"
}

main "$@"
