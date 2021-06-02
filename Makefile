.PHONY: all bin proto

all: bin

bin:
	./build.sh $(type) $(target)

proto:
	protoc-gen build
