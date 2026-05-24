#!/bin/bash

set -ex

cp -R /work/src/shanhu.io/homedrv/drv-ts .
(cd drv-ts; npm ci)
(cd drv-ts; make dist)
