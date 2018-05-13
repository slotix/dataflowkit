FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir /dfk

ADD testserver /dfk/testserver
WORKDIR /dfk
EXPOSE 12345
ENTRYPOINT /dfk/testserver