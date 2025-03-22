# libxivpn

[Xray](https://github.com/xtls/xray-core) + [tun2socks](https://github.com/xjasonlyu/tun2socks)

## How to build

- Install android NDK `27.0.12077973`
- Install golang
- Install `upx`
- Set environment variable `NDK` to the location of your toolchain (for example `/home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64`)
- Run `./build.sh`

## Vscode example configuration
`.vscode/c_cpp_properties.json`

```json
{
    "configurations": [
        {
            "name": "NDK",
            "includePath": [
                "${workspaceFolder}/**",
                "/home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64/sysroot/usr/include/**"
            ],
            "defines": [],
            "compilerPath": "/home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang",
            "cStandard": "c17",
            "cppStandard": "c++17",
            "intelliSenseMode": "linux-clang-x64"
        }
    ],
    "version": 4
}
```

`.vscode/settings.json`

```json
{
    "go.toolsEnvVars": {
        "CGO_CFLAGS": "-target x86_64-linux-android",
        "CXX": "/home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang++",
        "CC": "/home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang",
        "TARGET": "x86_64-linux-android",
        "GOARCH": "amd64",
        "GOOS": "android",
        "CGO_ENABLED": "1",
    }
}
```