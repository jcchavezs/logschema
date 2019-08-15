.PHONY: build

run:
	go run cmd/logschema/main.go

build:
	go build -o build/logschema cmd/logschema/main.go