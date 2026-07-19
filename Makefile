.PHONY: run test vet demo build

build:
	mkdir -p bin
	go build -o bin/relay ./cmd/relay

run:
	go run ./cmd/relay

test:
	go test ./... -count=1

vet:
	go vet ./...

demo:
	./scripts/demo.sh
