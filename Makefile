.PHONY: lint test deploy

lint:
	golangci-lint run ./...

test:
	go test ./...

deploy:
	gcloud app deploy
