#/usr/bin/env bash

spider_dir="./spiders"
target_dir="./releases"

go_arch="amd64"
go_os="linux"

build_time=`date +%Y%m%d.%H:%M:%S`
build_commit=`git rev-parse --short=12 HEAD`
build_branch=`git branch --show-current`
clipkg="github.com/voiladev/go-crawler/pkg/cli"
ldflags="-X $clipkg.buildTime=$build_time -X $clipkg.buildCommit=$build_commit -X $clipkg.buildBranch=$build_branch"

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
            echo BUILD $root/$sub/
            name=$domain.$sub.bin
            go vet $root/$sub/ && GOOS=$go_os GOARCH=$go_arch go build -ldflags "$ldflags -X $clipkg.buildName=$name" -o $target_dir/$name $root/$sub/ || exit 1
        fi

        for sub2 in `ls "$root/$sub"`; do
            if [ ! -d "$root/$sub/$sub2" ]; then
                continue
            fi

            if [ -f "$root/$sub/$sub2/main.go" ]; then
                echo BUILD $root/$sub/$sub2
                name=$domain.$sub.$sub2.bin
                go vet $root/$sub/$sub2 && GOOS=$go_os GOARCH=$go_arch go build -ldflags "$ldflags" -o $target_dir/$name $root/$sub/$sub2 || exit 1
            fi
        done
    done
done
