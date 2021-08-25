#!/usr/bin/env bash

typ=""
target="$2"

if [ "$1" = "go" ] || [ "$1" = "py" ]; then
    typ="$1"
else
    target="$1"
fi

spider_dir="./bin"
target_dir="./releases"

build_time=`date +%Y%m%d.%H:%M:%S`
build_commit=`git rev-parse --short=12 HEAD`
build_branch=`git branch --show-current`

if [ "$(uname -s|grep Darwin)" = "Darwin" ] && [ "$(uname -m)" = "arm64" ]; then
    # mac m1
    go_os="darwin"
    go_arch="arm64"
else
    go_os="linux"
    go_arch="amd64"
fi
go_clipkg="github.com/voiladev/VoilaCrawler/pkg/cli"
go_ldflags="-X $go_clipkg.buildTime=$build_time -X $go_clipkg.buildCommit=$build_commit -X $go_clipkg.buildBranch=$build_branch"

py_src_dir="./src"
py_loglevel="ERROR"

# buildGo generate executable binary
# params: $1=path, $2=targetname
buildGo() {
    path=$1
    name=$2
    echo "GO BUILD $path => $target_dir/$name"
    go vet $path && GOOS=$go_os GOARCH=$go_arch go build -ldflags "$go_ldflags -X $go_clipkg.buildName=$name" -o $target_dir/$name $path || exit 1
}

# buildPy generate executable binary with pyinstaller
# params: $1=sourcefile $2=targetname
buildPy() {
    return
    path=$1
    name=$2
    echo "PY BUILD $path => $target_dir/$name"
    pyinstaller --onefile --paths $py_src_dir --distpath=$target_dir --workpath=/tmp/pycrawler --specpath=/tmp/pycrawler -n $name --log-level=$py_loglevel $path
}

build() {
    path=$1
    name=$2
    if [ "$typ" = "go" ] || [ "$typ" = "" ]; then
        if [ -f "$path/main.go" ]; then
            buildGo "$path" "$name"
        fi
    fi

    if [ "$typ" = "py" ] || [ "$typ" = "" ]; then
        if [ -f "$path/main.py" ]; then
            buildPy "$path/main.py" "$name.py"
        fi
    fi
}

if [ -z "$target" ];  then
    rm -rf $target_dir 
    
    for domain in `ls $spider_dir`; do
        root="$spider_dir/$domain"
        if [ ! -d $root ]; then
            continue
        fi
        for sub in `ls $root`; do
            if [ ! -d "$root/$sub" ]; then
                continue
            fi

            build "$root/$sub" "$domain.$sub.bin"
            for sub2 in `ls "$root/$sub"`; do
                if [ ! -d "$root/$sub/$sub2" ]; then
                    continue
                fi

                build "$root/$sub/$sub2" "$domain.$sub.$sub2.bin"
            done
        done
    done
else
    if [ -d $spider_dir/$target ]; then
        name="${target//\//.}"
        build "$spider_dir/$target" "$name.bin"
    fi
fi
