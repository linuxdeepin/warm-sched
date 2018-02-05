all:

build-ctrl:
	cd ctl && go build -o ../warmctl

build-daemon:
	cd daemon && go build -o ../warm-daemon

clean:
	rm -rf warmctl warm-daemon

test:
	cd daemon && go test
	cd core && go test
	cd eventsource && go test

run-daemon: build-daemon
	./warm-daemon
