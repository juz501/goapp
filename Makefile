PACKAGES= github.com/urfave/negroni github.com/unrolled/render github.com/juz501/go_logger_middleware github.com/juz501/go_static_middleware

all: build run


build: clean
	GOPATH=`pwd -P` go build -o $(PWD)/bin/goapp goapp.go

clean:
	rm -f bin/goapp

install: uninstall 
	GOPATH=`pwd -P` go get ${PACKAGES}

uninstall:
	rm -rf src/github.com

run: build
	GOBASEROUTE=goapp $(PWD)/bin/goapp

# the following are used for setting up a daemon via systemd
# systemd service script not included

daemon: build restart

cleansymlink:
	sudo rm -f /usr/sbin/goapp

symlink: cleansymlink
	sudo ln -s $(PWD)/bin/goapp /usr/sbin/goapp

add: symlink 
	sudo cp -f systemd/system/goapp.service /etc/systemd/system
	sudo systemctl daemon-reload

remove: disable
	sudo rm -f /etc/systemd/system/goapp.service

enable: add
	sudo systemctl enable goapp

start:
	sudo systemctl start goapp

stop:
	sudo systemctl stop goapp

restart:
	sudo systemctl restart goapp

disable:
	sudo systemctl disable goapp
