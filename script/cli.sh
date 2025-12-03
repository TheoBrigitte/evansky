#!/usr/bin/env bash
#
# Author: Théo Brigitte
# Date: 2025-12-03

# Usage: cli.sh [options] [args...]
#
# This script interacts with The Movie Database (TMDb) API to search for media items.
# Arguments are passed as URL-encoded query parameters and are expected in the format key=value.
#
# Options:
#   -h, --help            Show this help message and exit
#       --type            Media type to search (movie, tv, multi). Default: movie
#
# Examples:
#   cli.sh "query=Alexandre" year=2004 language=en --type movie
#   cli.sh "query=Star Wars" year=1977 language=en --type collection
set -eu

BIN="$(basename "$0")"
MEDIA_TYPE="movie"

# Log variables and functions
RED='\033[0;31m'
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
BLUE='\033[0;34m'
NO_COLOR='\033[0m'

echo_stderr() {
  echo -e "$@" 1>&2
}

ok() {
  echo_stderr "${GREEN}OK${NO_COLOR}    $1"
}

info() {
  echo_stderr "${BLUE}INFO${NO_COLOR}  $1"
}

warn() {
  echo_stderr "${ORANGE}WARN${NO_COLOR}  $1"
}

error() {
  echo_stderr "${RED}ERROR${NO_COLOR} $1"
  exit 1
}

# print help message for a specific command
print_help() {
  sed -Ene '/#\s?Usage: '"$BIN $*"'/,/^([^#]|$)/{p; /^([^#]|$)/q}' "$0" | sed -e '$d; s/#\s\?//'
  exit
}

main() {
  # Arguments variables
  # Process arguments
  local args=()
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        # Display help message and exit.
        print_help;;
      -t|--type)
        # Take the next argument as the file name
        MEDIA_TYPE="${2-}"
        [[ -z "$MEDIA_TYPE" ]] && error "$1 requires an argument"
        shift;;
      -?*)
        error "Unknown option $1";;
      *)
        # Positional argument - add to args array
        # This allows for arbitrary arguments positioning
        args+=("$1");;
    esac
    shift
  done

  # Reset positional parameters to remaining arguments.
  set -- "${args[@]}"

  args=()
  while [[ $# -gt 0 ]]; do
    args+=(--data-urlencode "$1")
    shift
  done

  set -x
  curl -sS --get --request GET \
    --url "https://api.themoviedb.org/3/search/$MEDIA_TYPE" \
    --header "Authorization: Bearer $TMDB_API_KEY_READ" \
    --header "Accept: application/json" \
    "${args[@]}"
}

main "$@"
