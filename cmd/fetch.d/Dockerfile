FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=slotix/dfk-binary /go/src/github.com/slotix/dataflowkit/cmd/fetch.d/fetch.d /
RUN ls
EXPOSE 8000
ENTRYPOINT ./fetch.d