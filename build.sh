#!/bin/bash

set -euo pipefail

set -x

lets build dockers/dockers
go install ./cmd/homerelease
homerelease build
