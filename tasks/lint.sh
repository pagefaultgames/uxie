#! /bin/bash

set -euo pipefail

go_files=()
for a in "$@"; do
    case "$a" in
        *.go) go_files+=("$a") ;;
    esac
done

if [ ${#go_files[@]} -eq 0 ]; then
    echo "No Go files to lint."
    exit 0
fi

if [ -z "$CI" ]; then
    golangci-lint fmt "${go_files[@]}"
fi
golangci-lint run --fix "${go_files[@]}"
