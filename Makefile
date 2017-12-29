all: clean update build test

PKG_VERS := github.com/jmyounker/vers

CMD := jx
PKG_NAME := jx

GOFMT=gofmt -s 
GOFILES=$(wildcard *.go)

clean:
	rm -rf $(CMD) target

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
ifeq ($(shell uname),Darwin)
	$(eval PREFIX := /usr/local)
	$(eval INSTALL_USER := $(shell id -u)) 
	$(eval INSTALL_GROUP := $(shell id -g)) 
else
	$(eval PREFIX := /usr)
	$(eval INSTALL_USER := root)
	$(eval INTALL_GROUP := root)
endif
endif

install: set-prefix build
	install -m 755 -o $(INSTALL_USER) -g $(INSTALL_GROUP) $(CMD) $(PREFIX)/bin/$(CMD)

format:
	$(GOFMT) -w $(GOFILES)

package-base: test
	mkdir target
	mkdir target/model
	mkdir target/package

package-osx: set-version package-base
	mkdir target/model/osx
	mkdir target/model/osx/usr
	mkdir target/model/osx/usr/local
	mkdir target/model/osx/usr/local/bin
	install -m 755 $(CMD) target/model/osx/usr/local/bin/$(CMD)
	fpm -s dir -t osxpkg -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/osx .

package-rpm: set-version package-base
	mkdir target/model/linux-x86-rpm
	mkdir target/model/linux-x86-rpm/usr
	mkdir target/model/linux-x86-rpm/usr/bin
	install -m 755 $(CMD) target/model/linux-x86-rpm/usr/bin/$(CMD)
	fpm -s dir -t rpm -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-rpm .

package-deb: set-version package-base
	mkdir target/model/linux-x86-deb
	mkdir target/model/linux-x86-deb/usr
	mkdir target/model/linux-x86-deb/usr/bin
	install -m 755 $(CMD) target/model/linux-x86-deb/usr/bin/$(CMD)
	fpm -s dir -t deb -n $(PKG_NAME) -v $(VERSION) -p target/package -C target/model/linux-x86-deb .

