all: clean update build test

PKG_VERS := github.com/jmyounker/vers

clean:
	rm -rf jx target

update:
	go get
	go get $(PKG_VERS)

build-vers:
	make -C $$GOPATH/src/$(PKG_VERS) build

set-version: build-vers
	$(eval VERSION := $(shell $$GOPATH/src/$(PKG_VERS)/vers -f version.json show))
	
build: set-version
	go build -ldflags "-X main.version=$(VERSION)"

test: build
	go test

set-prefix:
ifndef PREFIX
	$(eval PREFIX := /usr)
endif


install: set-prefix
	install -m 755 -o root -g wheel jx $(PREFIX)/bin/jx

package-base: test
	mkdir target
	mkdir target/model
	mkdir target/package

package-osx: set-version package-base
	mkdir target/model/osx
	mkdir target/model/osx/usr
	mkdir target/model/osx/usr/local
	mkdir target/model/osx/usr/local/bin
	install -m 755 jx target/model/osx/usr/local/bin/jx
	fpm -s dir -t osxpkg -n jx -v $(VERSION) -p target/package -C target/model/osx .

package-rpm: set-version package-base
	mkdir target/model/linux-x86-rpm
	mkdir target/model/linux-x86-rpm/usr
	mkdir target/model/linux-x86-rpm/usr/bin
	install -m 755 jx target/model/linux-x86-rpm/usr/bin/jx
	fpm -s dir -t rpm -n jx -v $(VERSION) -p target/package -C target/model/linux-x86-rpm .

package-deb: set-version package-base
	mkdir target/model/linux-x86-deb
	mkdir target/model/linux-x86-deb/usr
	mkdir target/model/linux-x86-deb/usr/bin
	install -m 755 jx target/model/linux-x86-deb/usr/bin/jx
	fpm -s dir -t deb -n jx -v $(VERSION) -p target/package -C target/model/linux-x86-deb .

