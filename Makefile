.PHONY: tidy vet test web-build build-linux build-windows pkg-rpm pkg-deb pkg

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

web-build:
	cd frontend && npm ci --no-audit --no-fund && npm run build

build-linux: web-build
	GOOS=linux GOARCH=amd64 go build -o bin/mitt-linux ./cmd/mitt

build-windows: web-build
	GOOS=windows GOARCH=amd64 go build -o bin/mitt-windows.exe ./cmd/mitt

pkg-rpm: build-linux
	mkdir -p dist
	nfpm package --config packaging/nfpm.yaml --packager rpm --target dist/

pkg-deb: build-linux
	mkdir -p dist
	nfpm package --config packaging/nfpm.yaml --packager deb --target dist/mitt_0.1.3_amd64.deb

pkg: pkg-rpm pkg-deb
