.PHONY: lint test test-js deploy

lint:
	golangci-lint run ./...

test: test-js
	go test ./...

test-js:
	node --test data/app/src/direction.test.js data/app/src/touch.test.js

deploy:
	gcloud app deploy
