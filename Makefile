FORMATS=$(shell go run . --formats 2>&1 | grep 'has file' | awk '{print $$3}')
FORMATEXES=$(patsubst %,%q,$(FORMATS))
ALLEXES=anyq $(FORMATEXES)
PREFIX=/usr/local

all:
	CGO_ENABLED=0 GOOS=linux go build -o anyq .
	for THING in $(FORMATEXES); do ln -sf anyq $$THING; done

install: all
	mkdir -p $$DESTDIR$$PREFIX/bin
	cp -a $(ALLEXES) $$DESTDIR$$PREFIX/bin/

clean:
	rm -f $(ALLEXES)
