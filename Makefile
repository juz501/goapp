PACKAGES= github.com/urfave/negroni github.com/unrolled/render github.com/juz501/go_logger_middleware github.com/juz501/go_static_middleware

all: build run

build:
	GOPATH=`pwd -P` go build -o bin/server server.go

install: clean
	GOPATH=`pwd -P` go get ${PACKAGES}

clean:
	rm -rf src/github.com

run:
	GOPATH=`pwd -P` PORT=8080 GOBASEROUTE=goapp bin/server
