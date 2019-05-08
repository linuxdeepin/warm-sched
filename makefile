export GOPATH=$(shell pwd)/vendor
all: build-ctrl build-daemon

build-ctrl:
	cd ctl && go build -o ../bin/warmctl

build-daemon:
	cd daemon && go build -o ../bin/warm-daemon

clean:
	rm -rf bin/warmctl bin/warm-daemon

UNAME_M := $(shell uname -m)
ifneq ($(UNAME_M),sw_64)
test:
	cd daemon && go test
	cd core && go test
	cd events && go test
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
