.PHONY: tidy vet test build-linux build-windows pkg-rpm pkg-deb pkg

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

pkg-rpm: build-linux
	mkdir -p dist
	nfpm package --config packaging/nfpm.yaml --packager rpm --target dist/

pkg-deb: build-linux
	mkdir -p dist
	nfpm package --config packaging/nfpm.yaml --packager deb --target dist/mitt_0.1.0_amd64.deb

pkg: pkg-rpm pkg-deb
