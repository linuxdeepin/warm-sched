all:
	cd ctl && go build -o ../warmctl
	cd daemon && go build -o ../warm-daemon

clean:
	rm -rf warmctl warm-daemon
