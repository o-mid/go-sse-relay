.PHONY: run test vet demo build

build:
	mkdir -p bin
	go build -mod=vendor -o bin/relay ./cmd/relay

run:
	go run -mod=vendor ./cmd/relay

test:
	go test -mod=vendor ./... -count=1

vet:
	go vet -mod=vendor ./...

demo:
	./scripts/demo.sh
