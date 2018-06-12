FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
# RUN mkdir /dfk
# ADD testserver /dfk/testserver
# WORKDIR /dfk
ADD testserver /testserver
ADD web /web
EXPOSE 12345
ENTRYPOINT /testserver
# ENTRYPOINT /dfk/testserver
