#!/bin/bash

set -ex

caco3 build dockers/dockers
go install ./cmd/homerelease
homerelease build
