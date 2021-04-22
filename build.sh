#/usr/bin/env bash

cd spiders || exit 1

target_dir="./releases"

for root in `ls`; do
    if [ ! -d $root ]; then
        continue
    fi
    for sub in `ls $root`; do
        if [ ! -d "$root/$sub" ]; then
            continue
        fi

        if [ -f "$root/$sub/main.go" ]; then
            echo go build -buildmode=plugin -o $target_dir/$root.$sub.so ./$root/$sub/
            go build -buildmode=plugin -o $target_dir/$root.$sub.so ./$root/$sub/ || exit 1
        fi

        for sub2 in `ls "$root/$sub"`; do
            if [ ! -d "$root/$sub/$sub2" ]; then
                continue
            fi

            if [ -f "$root/$sub/$sub2/main.go" ]; then
                echo go build -buildmode=plugin -o $target_dir/plugins/$root.$sub.$sub2.so ./$root/$sub/$sub2
                go build -buildmode=plugin -o $target_dir/$root.$sub.$sub2.so ./$root/$sub/$sub2 || exit 1
            fi
        done
    done
done
