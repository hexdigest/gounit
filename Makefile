all: test lint

lint:
	golint ./... && go vet ./...

test:
	go test -race ./...

install:
	go install ./cmd/gounit

all: lint test install
