all: gen build

build: gw

gen:

gw: cmd/gw/main.go
	go build ./cmd/gw

test:
	go test -v -cover ./...


clean:
	-rm gw


.PHONY: gen clean



