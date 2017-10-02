HANDLER=handler
PACKAGE=package

# entrypoints
build: clean deps
	docker-compose run --rm sls-go make _build
.PHONY: build

test: deps
	docker-compose run --rm sls-go make _test
.PHONY: test

deploy:
	docker-compose run --rm sls-go make _deploy
.PHONY: deploy

remove:
	docker-compose run --rm sls-go make _remove
.PHONY: remove

shell:
	docker-compose run --rm sls-go bash
.PHONY: shell

# helpers
clean:
	docker-compose run --rm sls-go make _clean
.PHONY: clean

deps:
	docker-compose run --rm sls-go make _deps
.PHONY: deps

# target to run within container
_clean:
	@rm -rf .serverless $(HANDLER).zip $(HANDLER).so $(PACKAGE).zip

_deps:
	@dep ensure

_build:
	@go build -buildmode=plugin -ldflags='-w -s' -o $(HANDLER).so
	@pack $(HANDLER) $(HANDLER).so $(PACKAGE).zip
	@chown $(shell stat -c '%u:%g' .) $(HANDLER).so $(PACKAGE).zip

_test:
	@go test

_deploy:
	sls deploy -v

_remove:
	sls remove -v
