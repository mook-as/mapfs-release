#!/bin/bash -eux

pushd ~/workspace/mapfs-release
    bosh reset-release
popd

docker run \
-t \
-i \
--privileged \
-e DEV=TRUE \
-v ~/workspace/mapfs-release:/mapfs-release \
--workdir=/ \
bosh/main-bosh-docker \
/mapfs-release/scripts/run-bosh-release-tests-in-docker-env.sh
