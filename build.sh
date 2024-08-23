#!/bin/bash

rm libxivpn.so

echo $NDK # Example: /home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64

export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

export CC=$NDK/bin/aarch64-linux-android21-clang
export CXX=$NDK/bin/aarch64-linux-android21-clang++
export TARGET=aarch64-linux-android
export AR=$NDK/bin/llvm-ar
export LD=$NDK/bin/ld
export RANLIB=$NDK/bin/llvm-ranlib
export STRIP=$TOONDKLCHAIN/bin/llvm-strip

export CGO_CFLAGS="-target aarch64-linux-android"
export CGO_LDFLAGS="-v"

go build -buildmode=c-shared -trimpath -v -o libxivpn.so -ldflags="-s -w -buildid="
