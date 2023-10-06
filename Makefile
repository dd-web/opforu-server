build:
	@go build -o bin/bin cmd/app/main.go

run: build
	@./bin/bin

test:
	@go test -v ./...