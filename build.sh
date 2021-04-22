#/usr/bin/env bash

spider_dir="./spiders"
target_dir="./releases"

for domain in `ls $spider_dir`; do
    root="$spider_dir/$domain"
    if [ ! -d $root ]; then
        continue
    fi
    for sub in `ls $root`; do
        if [ ! -d "$root/$sub" ]; then
            continue
        fi

        if [ -f "$root/$sub/main.go" ]; then
            echo go build -buildmode=plugin -o $target_dir/$domain.$sub.so $root/$sub/
            go build -buildmode=plugin -o $target_dir/$domain.$sub.so $root/$sub/ || exit 1
        fi

        for sub2 in `ls "$root/$sub"`; do
            if [ ! -d "$root/$sub/$sub2" ]; then
                continue
            fi

            if [ -f "$root/$sub/$sub2/main.go" ]; then
                echo go build -buildmode=plugin -o $target_dir/$domain.$sub.$sub2.so $root/$sub/$sub2
                go build -buildmode=plugin -o $target_dir/$domain.$sub.$sub2.so $root/$sub/$sub2 || exit 1
            fi
        done
    done
done
