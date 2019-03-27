APP?=parse.d
GOOS?=linux
GOARCH?=amd64
PORT?=8002
IMAGE?=slotix/dfk-parse
RELEASE?=1.0.0
LATEST?=latest
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')


clean:
	rm -f ${APP}

build: clean
	CGO_ENABLED=0 \
	GOOS=${GOOS} GOARCH=${GOARCH} \
	go build \
        -ldflags "-s -w -X main.Release=${RELEASE} \
        -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}" \
		-a -installsuffix cgo \
        -o ${APP}

container: build
	docker build -t $(IMAGE):$(RELEASE) .
	docker tag $(IMAGE):$(RELEASE) ${IMAGE}:$(LATEST)


push: container
	docker push $(IMAGE):$(RELEASE)
	docker push ${IMAGE}:$(LATEST)

login:
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
	

# run: build
#     PORT=${PORT} ./${APP}

# test:
#     go test -v -race ./...