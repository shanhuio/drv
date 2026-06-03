#!/bin/bash

set -ex

cp -R /work .
(cd work; npm ci)
(cd work; make dist)
