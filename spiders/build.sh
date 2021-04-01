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

        if [ -f "$root/$sub/main.go" ]; then
            echo go build -buildmode=plugin -o ../releases/plugins/$root.$sub.so ./$root/$sub/
            go build -buildmode=plugin -o ../releases/plugins/$root.$sub.so ./$root/$sub/
        fi

        for sub2 in `ls $root/$sub`; do
            if [ ! -d $root/$sub/$sub2 ]; then
                continue
            fi

            if [ -f "$root/$sub/main.go" ]; then
                echo go build -buildmode=plugin -o ../releases/plugins/$root.$sub.$sub2.so ./$root/$sub/$sub2
                go build -buildmode=plugin -o ../releases/plugins/$root.$sub.$sub2.so ./$root/$sub/$sub2
            fi
        done
    done
done
