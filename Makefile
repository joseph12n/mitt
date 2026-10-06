.PHONY: tidy vet test build-linux build-windows

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/mitt-linux ./cmd/mitt

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/mitt-windows.exe ./cmd/mitt
