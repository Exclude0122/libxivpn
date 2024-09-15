package main

/*

#include "libxivpn_jni.h"

*/
import "C"
import (
	"fmt"
	"strings"

	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all"
)

var xrayServer core.Server

func log(msg string) {
	cString := C.CString(msg)
	C.libxivpn_log(cString)
}

//export libxivpn_version
func libxivpn_version() *C.char {
	return C.CString(strings.Join(core.VersionStatement(), "\n"))
}

//export libxivpn_start
func libxivpn_start(cConfig *C.char, socksPort C.int, fd C.int) *C.char {
	config := C.GoString(cConfig)

	log(fmt.Sprintln("start", config, socksPort, fd))

	defer func() {
		r := recover()
		if r != nil {
			log(fmt.Sprintf("panic: %v", r))
		}
	}()

	// xray

	xrayConfig, err := core.LoadConfig("json", strings.NewReader(config))
	if err != nil {
		log("libxivpn_start xray1: " + err.Error())
		return C.CString("libxivpn_start load config: " + err.Error())
	}

	xrayServer, err = core.New(xrayConfig)
	if err != nil {
		log("libxivpn_start xray2: " + err.Error())
		return C.CString("libxivpn_start create core: " + err.Error())
	}

	err = xrayServer.Start()
	if err != nil {
		log("libxivpn_start xray3: " + err.Error())
		return C.CString("libxivpn_start start core: " + err.Error())
	}

	// tun2socks
	engine.Insert(&engine.Key{
		MTU:        1400,
		Proxy:      "socks5://10.89.64.1:" + strconv.Itoa(int(socksPort)),
		Device:     "fd://" + strconv.Itoa(int(fd)),
		LogLevel:   "debug",
		UDPTimeout: time.Second * 10,
	})
	engine.Start()

	log(fmt.Sprintln("started"))

	return C.CString("")
}

//export libxivpn_stop
func libxivpn_stop() {
	defer func() {
		r := recover()
		if r != nil {
			log(fmt.Sprintf("panic: %v", r))
		}
	}()

	log("stop")
	engine.Stop()
	if xrayServer != nil {
		xrayServer.Close()
		xrayServer = nil
	}
}

func init() {
	log("init")

	go func() {
		r, w, _ := os.Pipe()
		os.Stdout = w
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			log("stdout: " + scanner.Text())
		}
	}()
	go func() {
		r, w, _ := os.Pipe()
		os.Stderr = w
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			log("stderr: " + scanner.Text())
		}
	}()

}

func main() {

}
