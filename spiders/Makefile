.PHONY: all

all: 
	go vet ./...
	go build -buildmode=plugin -o ../releases/plugins/com.asos.so ./com/asos/
	go build -buildmode=plugin -o ../releases/plugins/com.footlocker.so ./com/footlocker/
	go build -buildmode=plugin -o ../releases/plugins/com.nordstrom.so ./com/nordstrom/
	go build -buildmode=plugin -o ../releases/plugins/com.ruelala.so ./com/ruelala/
	go build -buildmode=plugin -o ../releases/plugins/com.tiktok.so ./com/tiktok/
	go build -buildmode=plugin -o ../releases/plugins/com.topshop.so ./com/topshop/
	go build -buildmode=plugin -o ../releases/plugins/com.urbanoutfitters.so ./com/urbanoutfitters/
