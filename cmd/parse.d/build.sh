#!/bin/sh
env GOOS=linux go build -v 
docker build --force-rm -t slotix/dfk-parse .