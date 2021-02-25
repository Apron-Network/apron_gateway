all: gen build

build: gw

SOURCES = $(wildcard internal/*/*.go internal/*.go cmd/*/*.go)


gen:

gw: $(SOURCES)
	go build ./cmd/gw

test:
	go test -v -cover ./...


clean:
	-rm gw


.PHONY: gen clean



