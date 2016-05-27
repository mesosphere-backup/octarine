
all: build

build:
	gox -arch=amd64 -os="linux darwin windows"
