#!/usr/bin/env bash
docker build --rm -f Dockerfile-binary -t 'slotix/dfk-binary' .
docker build --rm -f cmd/fetch.d/Dockerfile -t 'slotix/dfk-fetch' .
docker build --rm -f cmd/parse.d/Dockerfile -t 'slotix/dfk-parse' .
docker build --rm -f testserver/Dockerfile -t 'slotix/dfk-testserver' .
#docker build --rm -f webserver/Dockerfile -t 'slotix/dfk-webserver' .

#docker push slotix/dfk-fetch
#docker push slotix/dfk-parse
#docker push slotix/dfk-testserver
#docker push slotix/dfk-webserver