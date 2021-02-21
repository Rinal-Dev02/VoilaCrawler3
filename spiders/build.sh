#/usr/bin/env bash

echo go vet ./...
go vet ./...

for root in `ls`; do
    if [ ! -d $root ]; then
        continue
    fi
    for sub in `ls $root`; do
        if [ ! -d "$root/$sub" ]; then
            continue
        fi
        echo go build -buildmode=plugin -o ../releases/plugins/$root.$sub.so ./$root/$sub/
        go build -buildmode=plugin -o ../releases/plugins/$root.$sub.so ./$root/$sub/
    done
done
