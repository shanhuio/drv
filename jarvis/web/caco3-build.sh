#!/bin/bash

set -ex

cp -R /work/jarvis/web .
(cd web; npm ci)
(cd web; make dist)
