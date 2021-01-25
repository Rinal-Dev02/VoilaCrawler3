.PHONY: all

all: 
	go vet ./...
	go build -buildmode=plugin -o ../releases/plugins/com.ruelala.so ./com/ruelala/
	go build -buildmode=plugin -o ../releases/plugins/com.asos.so ./com/asos/
