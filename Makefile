build:
	@go build -o bin/dcpoker

run: build
	@./bin/dcpoker

test:
	go test -v ./...