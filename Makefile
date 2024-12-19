TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=github.com
NAMESPACE=ably
NAME=ably
BINARY=terraform-provider-${NAME}
VERSION=0.8.0
OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH_NATIVE := $(shell uname -m)
ARCH_MAPPED := $(shell echo "$(ARCH_NATIVE)" | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/' -e 's/armv7l/arm/')
ifneq ($(ARCH_MAPPED),$(filter amd64 arm64 arm,$(ARCH_MAPPED)))
$(error Unsupported architecture: $(ARCH_NATIVE))
endif
ARCH ?= $(ARCH_MAPPED) 
OS_ARCH=${OS}_${ARCH}

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_darwin_arm64
	GOOS=freebsd GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux   GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux   GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux   GOARCH=arm64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_arm64
	GOOS=linux   GOARCH=arm   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386   go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.VERSION=${VERSION}" -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=5m -parallel=10

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
