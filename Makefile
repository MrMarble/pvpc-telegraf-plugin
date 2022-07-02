.DEFAULT_GOAL := build

build:
	go build -o bin/pvpc cmd/main.go

deps:
	go mod vendor

clean:
	rm -r bin

test:
	go test -timeout 30s -count=1 ./plugins/inputs/pvpc

run:
	make build
	./bin/pvpc --config plugin.conf
