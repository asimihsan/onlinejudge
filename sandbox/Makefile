# g++ -std=c++11 -Wall -Werror sandbox.cpp -lseccomp -o sandbox

PREFIX=/usr/local
CC=g++
CFLAGS+=-std=c++11 -Wall -Wextra -O2 -fPIE \
    -fstack-protector-all \
    $(shell pkg-config --cflags gio-2.0)
LDLIBS+=$(shell pkg-config --libs libseccomp) \
    $(shell pkg-config --libs gio-2.0) -lcap
LDFLAGS+=-pie -Wl,--as-needed,-z,relro,-z,now
SOURCES=sandbox.cpp

all: sandbox

sandbox: syscall-names.h
	$(CC) $(CFLAGS) $(SOURCES) $(LDLIBS) $(LDFLAGS) -o sandbox
	chown nobody:nogroup sandbox
	chmod +s sandbox

syscall-names.h: /usr/include/x86_64-linux-gnu/asm/unistd_64.h
	echo "static const char *syscall_names[] = {" > $@ ;\
	cat /usr/include/x86_64-linux-gnu/asm/unistd_64.h | \
		LC_ALL=C sed -r -n -e 's/^\#define[ \t]+__NR_([a-z0-9_]+)[ \t]+([0-9]+)(.*)/"\1",/p' >> $@ ;\
	echo "};" >> $@

install: sandbox
	install -Dm755 $< $(DESTDIR)$(PREFIX)/bin/$<
	chown nobody:nogroup $(DESTDIR)$(PREFIX)/bin/$<
	chmod +s $(DESTDIR)$(PREFIX)/bin/$<

clean:
	rm -f sandbox syscall-names.h
