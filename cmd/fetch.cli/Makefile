APP?=fetch.cli
RELEASE?=0.9.0
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')


clean:
	rm -f ${APP}

build: clean
	CGO_ENABLED=0 \
	GOOS=linux \
	go build \
        -ldflags "-s -w -X main.Release=${RELEASE} \
        -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}" \
		-a -installsuffix cgo \
        -o ${APP}

# run: build
#     PORT=${PORT} ./${APP}

# test:
#     go test -v -race ./...