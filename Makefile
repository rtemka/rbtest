all: build test lint

deps:
	go mod tidy && go mod download

build: deps
	go build -o ./bin/ ./cmd/

run:
	go run ./cmd/

test:
	go test -v -cover -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf ./bin