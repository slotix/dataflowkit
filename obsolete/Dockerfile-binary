# go build dfk binary files
FROM golang:latest as builder
ADD . /go/src/github.com/slotix/dataflowkit
WORKDIR /go/src/github.com/slotix/dataflowkit/cmd/fetch.d
RUN make build
WORKDIR /go/src/github.com/slotix/dataflowkit/cmd/parse.d
RUN make build
WORKDIR /go/src/github.com/slotix/dataflowkit/testserver
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o testserver .