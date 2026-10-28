dev:
	go run ./cmd/bot

fmt:
	find . -type f -name '*.go' -exec go tool gofumpt -w {} +
	find . -type f -name '*.go' -exec go tool goimports -w {} +
	find . -type f -name '*.go' -exec go tool golines -m 80 -w {} +

lint:
	golangci-lint run $(filter-out $@,$(MAKECMDGOALS))

build:
	rm -rf bin && go build -o bin/bot ./cmd/bot

start:
	./bin/bot