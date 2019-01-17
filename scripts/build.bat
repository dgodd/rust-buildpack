#!/usr/bin/env bash
set -exuo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

GOOS=windows go build -ldflags="-s -w" -o bin/supply.exe ./supply/cli
GOOS=windows go build -ldflags="-s -w" -o bin/finalize.exe ./finalize/cli
rm bin/{supply,finalize}
