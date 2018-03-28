FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir /dfk

ADD fetch.d /dfk/fetch.d
WORKDIR /dfk
EXPOSE 8000
ENTRYPOINT /dfk/fetch.d