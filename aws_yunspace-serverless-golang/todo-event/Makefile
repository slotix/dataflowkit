HANDLER=handler
PACKAGE=package
ifdef DOTENV
	DOTENV_TARGET=dotenv
else
	DOTENV_TARGET=.env
endif

########################################
# entrypoints - run within docker only #
########################################
deps:
	docker-compose run --rm sls-go make _deps
.PHONY: deps

build: $(DOTENV_TARGET) _clean deps
	docker-compose run --rm sls-go make _build
.PHONY: build

test: $(DOTENV_TARGET) deps
	docker-compose run --rm sls-go make _test
.PHONY: test

deploy: $(DOTENV_TARGET)
	docker-compose run --rm sls-go make _deploy
.PHONY: deploy

remove: $(DOTENV_TARGET)
	docker-compose run --rm sls-go make _remove
.PHONY: remove

shell: $(DOTENV_TARGET)
	docker-compose run --rm sls-go bash
.PHONY: shell

########################################
# others - run within or out of docker #
########################################
.env:
	@echo "Create .env with .env.template"
	cp .env.template .env

dotenv:
	@echo "Switching .env to $(DOTENV)"
	cp $(DOTENV) .env

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
	@go test $(shell go list ./... | grep -v /vendor/)

_deploy:
	@sls deploy -v

_remove:
	@sls remove -v
