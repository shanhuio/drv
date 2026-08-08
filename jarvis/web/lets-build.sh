#!/bin/bash

set -ex

cp -R /work .
(cd work; npm ci --allow-git=all)
(cd work; make dist)
