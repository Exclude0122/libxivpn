#!/bin/bash

# usage: ./build [all / arm64 / x86_64]

arch=$1
if [ "$arch" = "" ]
then
    arch="all"
fi

rm libxivpn_x86_64.so
rm libxivpn_arm64.so

echo $NDK # Example: /home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64

export GOOS=android
export CGO_ENABLED=1

export AR=$NDK/bin/llvm-ar
export LD=$NDK/bin/ld
export RANLIB=$NDK/bin/llvm-ranlib
export STRIP=$TOONDKLCHAIN/bin/llvm-strip

export CGO_LDFLAGS="-v"

# arm64
if [ "$arch" = "all" ] || [ "$arch" = "arm64" ]
then
    echo "Builing arm64"

    export CGO_CFLAGS="-target aarch64-linux-android"
    export GOARCH=arm64
    export CC=$NDK/bin/aarch64-linux-android21-clang
    export CXX=$NDK/bin/aarch64-linux-android21-clang++
    export TARGET=aarch64-linux-android

    go build -buildmode=c-shared -trimpath -o libxivpn_arm64.so -ldflags="-s -w -buildid=" -buildvcs=false

    # chmod +x libxivpn_arm64.so
    # upx --android-shlib libxivpn_arm64.so
fi

# x86_64
if [ "$arch" = "all" ] || [ "$arch" = "x86_64" ]
then
    echo "Builing x86_64"

    export CGO_CFLAGS="-target x86_64-linux-android"
    export GOARCH=amd64
    export CC=$NDK/bin/x86_64-linux-android21-clang
    export CXX=$NDK/bin/x86_64-linux-android21-clang++
    export TARGET=x86_64-linux-android

    go build -buildmode=c-shared -trimpath -o libxivpn_x86_64.so -ldflags="-s -w -buildid=" -buildvcs=false

    # chmod +x libxivpn_x86_64.so
    # upx --android-shlib libxivpn_x86_64.so
fi