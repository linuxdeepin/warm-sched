PREFIX = /usr
GOPATH_DIR = gopath
GOPKG_PREFIX = github.com/linuxdeepin/warm-sched
export GOPATH=$(shell go env GOPATH)
GOCMD = env GOPATH="$(CURDIR)/$(GOPATH_DIR):$(CURDIR)/vendor:$(GOPATH)" go
GOBUILD =$(GOCMD) build $(GO_BUILD_FLAGS)
GOTEST = $(GOCMD) test -v
export GO111MODULE=off
export GOCACHE=/tmp/

all: build-ctrl build-daemon build-snapshot-convert

prepare:
	@mkdir -p bin
	@mkdir -p $(GOPATH_DIR)/src/$(dir $(GOPKG_PREFIX));
	@[ -e  $(GOPATH_DIR)/src/$(GOPKG_PREFIX) ] || ln -snf ../../../.. $(GOPATH_DIR)/src/$(GOPKG_PREFIX);

build-ctrl: prepare
	$(GOBUILD) -o bin/warmctl $(GOPKG_PREFIX)/cmd/ctl

build-daemon: prepare
	$(GOBUILD) -o bin/warm-daemon $(GOPKG_PREFIX)/cmd/daemon

build-snapshot-convert: prepare
	$(GOBUILD) -o bin/warm-snapshot-convert $(GOPKG_PREFIX)/cmd/snapshot-convert

clean:
	rm -rf gopath
	rm -rf bin/warmctl bin/warm-daemon

UNAME_M := $(shell uname -m)
ifneq ($(UNAME_M),sw_64)
test:
	$(GOTEST) $(GOPKG_PREFIX)/cmd/daemon
	$(GOTEST) $(GOPKG_PREFIX)/cmd/ctl
	$(GOTEST) $(GOPKG_PREFIX)/core
	$(GOTEST) $(GOPKG_PREFIX)/events
else
test:
	true
endif

run-daemon: build-daemon
	./bin/warm-daemon -auto=false

t:
	debuild -uc -us
	scp ../warm-sched_0.2.0_amd64.deb deepin@pc:~
	ssh -t deepin@pc sudo dpkg -i /home/deepin/warm-sched_0.2.0_amd64.deb
	dpkg-buildpackage -Tclean
	ssh -t deepin@pc sudo reboot
