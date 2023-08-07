all:
	CGO_ENABLED=0 GOOS=linux go build -o anyq .
	ln -sf anyq jsonq

clean:
	rm -f anyq jsonq