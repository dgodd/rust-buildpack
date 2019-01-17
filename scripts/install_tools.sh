#!/bin/bash
set -euo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

if [ ! -f .bin/buildpack-packager ]; then
go install github.com/cloudfoundry/libbuildpack/packager/buildpack-packager
fi