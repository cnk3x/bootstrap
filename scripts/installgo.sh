#!/usr/bin/env bash

set -e

version=$(wget -qO - https://golang.google.cn/dl/ | grep -Eo 'go([0-9\.]+).linux-amd64' | head -1 | cut -c 3-10 | cut -d "." -f 1-3)

echo "version: ${version}"

os=`uname | tr 'A-Z' 'a-z'`

baseRoot=/usr/local/golang
dl=${baseRoot}/dl
goRoot=${baseRoot}/go

mkdir -p ${dl} ${goRoot}

cd ${baseRoot}

fn=go${version}.${os}-amd64.tar.gz
saveTo=${dl}/${fn}

echo download
curl https://dl.google.com/go/${fn} -o ${saveTo}

echo unzip
rm -rf ${goRoot}
tar zxf ${saveTo}

for bin in `ls ${goRoot}/bin`
do
	rm -f /usr/local/bin/${bin}
	ln -s -f ${goRoot}/bin/${bin} /usr/local/bin/${bin}
done

go version
