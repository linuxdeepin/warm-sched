all:

build-ctrl:
	cd ctl && go build -o ../warmctl

build-daemon:
	cd daemon && go build -o ../warm-daemon

clean:
	rm -rf warmctl warm-daemon

run-daemon: build-daemon
	rm /run/user/1000/warm-sched.socket
	./warm-daemon
