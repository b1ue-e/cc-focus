.PHONY: build install clean

BINARY = cc-focus
PREFIX ?= $(HOME)/.local

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(PREFIX)/bin
	cp $(BINARY) $(PREFIX)/bin/
	cp notify.sh $(PREFIX)/bin/cc-focus-notify

clean:
	rm -f $(BINARY)
