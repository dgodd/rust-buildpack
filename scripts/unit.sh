#!/usr/bin/env bash
# Runs the unit tests for this buildpack

set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

go test ./{supply,finalize}/{,cli}
