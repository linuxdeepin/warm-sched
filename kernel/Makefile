obj-m := test_module.o

KERNELDIR ?= /lib/modules/$(shell uname -r)/build
PWD := $(shell pwd)

all: build test

build:
	$(MAKE) -C $(KERNELDIR) M=$(PWD)

clean:
	rm -f *.o *~ core .depend .*.cmd *.ko *.mod.c
	rm -rf .tmp_versions
	rm Module.symvers
	rm modules.order

test:
	-@sudo rmmod test_module
	@sudo insmod ./test_module.ko
	cat /proc/snyh123 | less
