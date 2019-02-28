#!/usr/bin/env bash
#remove old docker images
docker rmi slotix/dfk-testserver
docker rmi slotix/dfk-parse
docker rmi slotix/dfk-fetch
docker rmi slotix/dfk-binary