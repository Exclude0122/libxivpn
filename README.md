# libxivpn

[Xray](https://github.com/xtls/xray-core) + [tun2socks](https://github.com/xjasonlyu/tun2socks)

## How to build

- Install android NDK `27.0.12077973`
- Install golang
- Install `upx`
- Set environment variable `NDK` to the location of your toolchain (for example `/home/USERNAME/Android/Sdk/ndk/27.0.12077973/toolchains/llvm/prebuilt/linux-x86_64`)
- Run `./build.sh all patch`

## IPC

Read `IPC.md`


## How to debug

0. Modify `build.sh` and temporarily remove `-buildmode=pie -trimpath` and `-ldflags="-s -w"`

1. Clone https://github.com/go-delve/delve

2. Open `pkg/proc/bininfo.go` and add `"android"` to switch conditions

```golang
func loadBinaryInfo(bi *BinaryInfo, image *Image, path string, entryPoint uint64) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	switch bi.GOOS {
	case "linux", "freebsd", "android": // append android here
		return loadBinaryInfoElf(bi, image, path, entryPoint, &wg)
	case "windows":
		return loadBinaryInfoPE(bi, image, path, entryPoint, &wg)
	case "darwin":
		return loadBinaryInfoMacho(bi, image, path, entryPoint, &wg)
	}
	return errors.New("unsupported operating system")
}
```


3. Build dlv

```bash
CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build ./cmd/dlv/
```

4. Copy `dlv` to `/data/user/0/io.github.exclude0122.xivpn/files/dlv` using ADB

5. Enter adb shell and execute the following commands

```bash
 # Note: in order to use run-as, the app must be debuggable
run-as io.github.exclude0122.xivpn
cd /data/user/0/io.github.exclude0122.xivpn/files
# Turn on XiVPN, and then find the PID of libxivpn.so
ps -A
# Attach to libxivpn.so. Change PID to the actual PID
./dlv attach PID --headless --listen "0.0.0.0:23456"
```

6. Start ADB forward

```bash
adb forward tcp:23456 tcp:23456
```

7. Connect to dlv using vscode. An example vscode `launch.json` is provided below:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Connect to server",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "port": 23456,
            "host": "127.0.0.1",
            "substitutePath": []
        }
    ]
}
```

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