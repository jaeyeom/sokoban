# Developer workflow for the Go server and the data/app frontend.
#
# check is the CI target: non-mutating format check and tests.
# lint is not part of check; CI does not install golangci-lint.
# test-js reads data/app/dist/bundle.js, so all finishes tests before build.

SERVER_BIN := /tmp/sokoban-webserver

.PHONY: all check \
	format check-format check-gofmt \
	lint fix \
	test test-go test-js \
	build build-go build-web \
	coverage coverage-html coverage-report clean-coverage \
	clean deploy

# Writers finish before readers. The webpack bundle is rewritten only after tests.
all: fix
	$(MAKE) check
	$(MAKE) build

check: check-format test

format:
	@files="$$(git ls-files '*.go')"; \
	if [ -n "$$files" ]; then \
		gofmt -w $$files; \
	fi

check-format:
	@unformatted="$$(gofmt -l $$(git ls-files '*.go'))"; \
	if [ -n "$$unformatted" ]; then \
		printf '%s\n' "$$unformatted"; \
		exit 1; \
	fi

check-gofmt: check-format

lint:
	golangci-lint run ./...

# format and fix both rewrite Go files, so fix runs only after format.
fix: format
	golangci-lint run --fix ./...

test: test-go test-js

test-go:
	go test ./...

test-js:
	npm --prefix data/app test

build: build-go build-web

build-go:
	go build -o $(SERVER_BIN) ./cmd/webserver

build-web:
	cd data/app && npx webpack --config webpack.config.js

coverage:
	go test -coverprofile=coverage.out ./...

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

coverage-report: coverage
	go tool cover -func=coverage.out

clean: clean-coverage
	rm -f $(SERVER_BIN)

clean-coverage:
	rm -f coverage.out coverage.html

# Rebuild the bundle App Engine serves, then deploy without prompting.
deploy: build-web
	gcloud app deploy --quiet
