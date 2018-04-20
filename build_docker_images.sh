#!/usr/bin/env bash

cd cmd/fetch.d && ./build.sh
cd ../parse.d && ./build.sh
#docker push slotix/dfk-fetch
#docker push slotix/dfk-parse