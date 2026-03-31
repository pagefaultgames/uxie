#! /usr/bin/env bash

set -eo pipefail

rm -rf test-results

# The file containing default `gotestsum` arguments.
args_file="task-args/test-args-local.txt"
if [ -n "$CI" ]; then
    args_file="task-args/test-args-ci.txt"
fi

declare -a base_args
readarray -t base_args < "$args_file"

# If no package identifiers are provided in local runs, default to "./...".
if [ -n "$CI" ]; then
    go tool gotest.tools/gotestsum "${base_args[@]}" "${cli_args[@]}"
    return
fi

has_pkg=false
for arg in "${cli_args[@]}"; do
    case "$arg" in
         ./...|./*)
            has_pkg=true
            break
            ;;
    esac
done

if ! $has_pkg; then
    echo "No package identifier found; defaulting to running everything..."
    cli_args+=("./...")
fi

go tool gotest.tools/gotestsum "${base_args[@]}" "$@"
