GOPATH=$(shell pwd)/vendor
all: update-vendor build-ctrl build-daemon

update-vendor:
	git submodule update --init

build-ctrl:
	cd ctl && go build -o ../bin/warmctl

build-daemon:
	cd daemon && go build -o ../bin/warm-daemon

clean:
	rm -rf bin/warmctl bin/warm-daemon

test:
	cd daemon && go test -race
	cd core && go test -race
	cd events && go test -race

run-daemon: build-daemon
	./bin/warm-daemon -auto=false

t:
	debuild -uc -us
	scp ../warm-sched_0.2.0_amd64.deb deepin@pc:~
	ssh -t deepin@pc sudo dpkg -i /home/deepin/warm-sched_0.2.0_amd64.deb
	ssh -t deepin@pc sudo reboot
