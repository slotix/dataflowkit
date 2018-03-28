FROM alpine:latest

RUN mkdir /dfk

ADD parse.d /dfk/parse.d
WORKDIR /dfk
EXPOSE 8001
ENTRYPOINT /dfk/parse.d