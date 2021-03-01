all: gen build

build: gw

SOURCES = $(wildcard internal/*/*.go internal/*.go cmd/*/*.go)


gen: proto/models.proto
	protoc --proto_path=proto --go_out=internal/models --go_opt=paths=source_relative proto/models.proto

gw: $(SOURCES)
	go build ./cmd/gw

test:
	go test -v -cover ./...


clean:
	-rm gw


.PHONY: gen clean



