.PHONY: proto

all: proto

proto:
	protoc -I. -I="${GOPATH}/src" -I="${GOPATH}/src/github.com/gogo/protobuf/protobuf" --gogofaster_out=. *.proto
