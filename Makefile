ifeq ($(GOPATH),)
export GOPATH=/tmp/go
endif
export PATH := $(PATH):$(GOPATH)/bin

INSTALL := /usr/bin/install
DBDIR := /var/run/redis/sonic-db/
GO ?= /usr/local/go/bin/go
TOP_DIR := $(abspath ..)
BUILD_DIR := build/bin
export GOBIN := $(abspath $(BUILD_DIR))
export PATH := $(PATH):$(GOBIN):$(shell dirname $(GO))

SRC_FILES=$(shell find . -name '*.go' | grep -v '_test.go' | grep -v '/tests/')

GO_DEPS := vendor/.done
PATCHES := $(wildcard patches/*.patch)
UNIT_TEST := $(shell if sudo [ -e /var/run/redis/redis.sock ]; then echo "exist"; else echo "noexist"; fi)

all: sonic-gnmi

go.mod:
	$(GO) mod init github.com/sonic-net/sonic-gnmi

$(GO_DEPS): go.mod $(PATCHES)
	$(GO) mod vendor
	$(GO) mod download golang.org/x/crypto@v0.0.0-20191206172530-e9b2fee46413
	$(GO) mod download github.com/jipanyang/gnxi@v0.0.0-20181221084354-f0a90cca6fd0
	cp -r $(GOPATH)/pkg/mod/golang.org/x/crypto@v0.0.0-20191206172530-e9b2fee46413/* vendor/golang.org/x/crypto/
	cp -r $(GOPATH)/pkg/mod/github.com/jipanyang/gnxi@v0.0.0-20181221084354-f0a90cca6fd0/* vendor/github.com/jipanyang/gnxi/
	./patches/apply.sh vendor
	chmod -R u+w vendor
	patch -d vendor -p0 < patches/gnmi_cli.all.patch
	patch -d vendor -p0 < patches/gnmi_set.patch
	patch -d vendor -p0 < patches/gnmi_get.patch
	patch -d vendor -p0 < patches/path.patch
	touch $@

go-deps: $(GO_DEPS)

go-deps-clean:
	$(RM) -r vendor

sonic-gnmi: $(GO_DEPS) libswss
ifeq ($(CROSS_BUILD_ENVIRON),y)
	$(GO) build -o ${GOBIN}/gnxi -mod=vendor github.com/sonic-net/sonic-gnmi/gnxi
	$(GO) build -o ${GOBIN}/gnmi_get -mod=vendor github.com/jipanyang/gnxi/gnmi_get
	$(GO) build -o ${GOBIN}/gnmi_set -mod=vendor github.com/jipanyang/gnxi/gnmi_set
	$(GO) build -o ${GOBIN}/gnmi_cli -mod=vendor github.com/openconfig/gnmi/cmd/gnmi_cli
	$(GO) build -o ${GOBIN}/gnoi_client -mod=vendor github.com/sonic-net/sonic-gnmi/gnoi_client
	$(GO) build -o ${GOBIN}/gnmi_dump -mod=vendor github.com/sonic-net/sonic-gnmi/gnmi_dump
else
	$(GO) install -mod=vendor github.com/sonic-net/sonic-gnmi/gnxi
	$(GO) install -mod=vendor github.com/jipanyang/gnxi/gnmi_get
	$(GO) install -mod=vendor github.com/jipanyang/gnxi/gnmi_set
	$(GO) install -mod=vendor github.com/openconfig/gnmi/cmd/gnmi_cli
	$(GO) install -mod=vendor github.com/sonic-net/sonic-gnmi/gnoi_client
	$(GO) install -mod=vendor github.com/sonic-net/sonic-gnmi/gnmi_dump
endif

# TODO: Create a new repo for this lib, sonic-restapi and sonic-gnmi can share this lib
libswss:
	make -C libcswsscommon
	sudo make -C libcswsscommon install

check:
ifeq ("$(UNIT_TEST)", "exist")
ifeq ($(wildcard ${DBDIR}/database_config.json),)
	sudo mkdir -p ${DBDIR}
	sudo cp ./testdata/database_config.json ${DBDIR}
endif
	sudo $(GO) test -coverprofile=coverage-gnmi.tmp -covermode=atomic -mod=vendor -v github.com/sonic-net/sonic-gnmi/gnmi_server -coverpkg ../...
	sudo $(GO) test -coverprofile=coverage-json.txt -covermode=atomic -mod=vendor -v github.com/sonic-net/sonic-gnmi/sonic_data_client
	sudo $(GO) test -coverprofile=coverage-dbus.txt -covermode=atomic -mod=vendor -v github.com/sonic-net/sonic-gnmi/sonic_service_client
	$(GO) get github.com/axw/gocov/...
	$(GO) get github.com/AlekSi/gocov-xml
	cat coverage-gnmi.tmp | grep -v "sonic-gnmi/proto" > coverage-gnmi.txt
	gocov convert coverage-*.txt | gocov-xml -source $(shell pwd) > coverage.xml
	rm -rf coverage-*.tmp
	rm -rf coverage-*.txt
endif

clean:
	$(RM) -r build
	$(RM) -r vendor

install:
	$(INSTALL) -D $(BUILD_DIR)/gnxi $(DESTDIR)/usr/sbin/gnxi
	$(INSTALL) -D $(BUILD_DIR)/gnmi_get $(DESTDIR)/usr/sbin/gnmi_get
	$(INSTALL) -D $(BUILD_DIR)/gnmi_set $(DESTDIR)/usr/sbin/gnmi_set
	$(INSTALL) -D $(BUILD_DIR)/gnmi_cli $(DESTDIR)/usr/sbin/gnmi_cli
	$(INSTALL) -D $(BUILD_DIR)/gnoi_client $(DESTDIR)/usr/sbin/gnoi_client
	$(INSTALL) -D $(BUILD_DIR)/gnmi_dump $(DESTDIR)/usr/sbin/gnmi_dump


deinstall:
	rm $(DESTDIR)/usr/sbin/gnxi
	rm $(DESTDIR)/usr/sbin/gnmi_get
	rm $(DESTDIR)/usr/sbin/gnmi_set
	rm $(DESTDIR)/usr/sbin/gnoi_client
	rm $(DESTDIR)/usr/sbin/gnmi_dump


