.PHONY: lint test test-js build check check-gofmt deploy

lint:
	golangci-lint run ./...

test: test-js
	go test ./...

test-js:
	node --test data/app/src/direction.test.js data/app/src/touch.test.js

check-gofmt:
	@test -z "$$(gofmt -l $$(git ls-files '*.go'))"

check: check-gofmt test

build:
	go build -o /tmp/sokoban-webserver ./cmd/webserver
	cd data/app && npx webpack --config webpack.config.js

deploy:
	gcloud app deploy
