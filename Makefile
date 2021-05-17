.PHONY: all bin

all: bin

proto:
	protoc-gen build

bin:
	./build.sh
