#/usr/bin/env bash

spider_dir="./bin"
src_dir="./src"
target_dir="./releases"

build_time=`date +%Y%m%d.%H:%M:%S`
build_commit=`git rev-parse --short=12 HEAD`
build_branch=`git branch --show-current`

go_arch="amd64"
go_os="linux"
go_clipkg="github.com/voiladev/VoilaCrawler/pkg/cli"
go_ldflags="-X $go_clipkg.buildTime=$build_time -X $go_clipkg.buildCommit=$build_commit -X $go_clipkg.buildBranch=$build_branch"

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
    path=$1
    name=$2
    echo "PY BUILD $path => $target_dir/$name"
    pyinstaller --onefile --paths $src_dir --distpath=$target_dir --workpath=/tmp/pycrawler --specpath=/tmp/pycrawler -n $name --log-level=$py_loglevel $path
}

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

        if [ -f "$root/$sub/main.go" ]; then
            buildGo "$root/$sub" "$domain.$sub.bin"
        fi
        if [ -f "$root/$sub/main.py" ]; then
            buildPy "$root/$sub/main.py" "$domain.$sub.bin.py"
        fi

        for sub2 in `ls "$root/$sub"`; do
            if [ ! -d "$root/$sub/$sub2" ]; then
                continue
            fi

            if [ -f "$root/$sub/$sub2/main.go" ]; then
                buildGo "$root/$sub/$sub2" "$domain.$sub.$sub2.bin"
            fi
            if [ -f "$root/$sub/$sub2/main.py" ]; then
                buildPy "$root/$sub/$sub2" "$domain.$sub.$sub2.bin.py"
            fi
        done
    done
done
