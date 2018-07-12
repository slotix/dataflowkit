FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=slotix/dfk-binary /go/src/github.com/slotix/dataflowkit/testserver/testserver .
ADD web /web
EXPOSE 12345
ENTRYPOINT testserver
