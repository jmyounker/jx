VERSION=0.9

.PHONY: clean

clean:
	go clean
	rm -rf usr *.pkg *.deb *.rpm

build: main.go
	go build -ldflags "-X main.Version=$(VERSION)"

pkg-deb: build
	rm -rf usr
	mkdir -p usr/local/bin
	cp -p jx usr/local/bin
	fpm -s dir -t deb -n jx -v $(VERSION) usr

pkg-osx: build
	rm -rf usr
	mkdir -p usr/local/bin
	cp -p jx usr/local/bin
	fpm -s dir -t osxpkg -n jx -v $(VERSION) usr

pkg-rpm: build
	rm -rf usr
	mkdir -p usr/local/bin
	cp -p jx usr/local/bin
	fpm -s dir -t rpm -n jx -v $(VERSION) usr

