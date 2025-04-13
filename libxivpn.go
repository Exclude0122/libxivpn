package main

/*
#include "libxivpn_jni.h"

*/
import "C"
import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all"
	"github.com/xtls/xray-core/transport/internet"
)

var xrayServer core.Server
var logFile *os.File
var registerControllerOnce sync.Once

func log(msg string) {
	cString := C.CString(msg)
	C.libxivpn_log(cString)

	if logFile != nil {

		_, err := logFile.WriteString("[" + time.Now().Format(time.DateTime) + "] " + msg + "\n")
		if err != nil {
			C.libxivpn_log(C.CString("write log file: " + err.Error()))
			logFile.Close()
			logFile = nil
			return
		}
		err = logFile.Sync()
		if err != nil {
			C.libxivpn_log(C.CString("sync log file: " + err.Error()))
			logFile.Close()
			logFile = nil
			return
		}
	}
}

//export libxivpn_version
func libxivpn_version() *C.char {
	return C.CString(strings.Join(core.VersionStatement(), "\n"))
}

//export libxivpn_start
func libxivpn_start(cConfig *C.char, socksPort C.int, fd C.int, cLogFile *C.char, cAsset *C.char) *C.char {
	config := C.GoString(cConfig)
	logFileName := C.GoString(cLogFile)
	asset := C.GoString(cAsset)

	log("set env XRAY_LOCATION_ASSET: " + asset)
	err := os.Setenv("XRAY_LOCATION_ASSET", asset)
	if err != nil {
		log("error set env: " + err.Error())
	}

	// log file
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		if err != nil {
			log("close log file 1: " + err.Error())
			return C.CString("close log file 1: " + err.Error())
		}
	}

	if logFileName != "" {
		var err error
		logFile, err = os.Create(logFileName)
		if err != nil {
			log("create log file: " + err.Error())
			return C.CString("create log file: " + err.Error())
		}
	}

	// register socket controller
	registerControllerOnce.Do(func() {
		log("register controller once")

		err := internet.RegisterDialerController(func(network, address string, conn syscall.RawConn) error {
			return conn.Control(func(fd uintptr) {
				log(fmt.Sprintf("protect dialer %s %s %d", network, address, fd))
				C.libxivpn_protect(C.int(fd))
			})
		})
		if err != nil {
			log("failed to register dialer controller")
		}

		err = internet.RegisterListenerController(func(network, address string, conn syscall.RawConn) error {
			return conn.Control(func(fd uintptr) {
				log(fmt.Sprintf("protect listener %s %s %d", network, address, fd))
				C.libxivpn_protect(C.int(fd))
			})
		})
		if err != nil {
			log("failed to register dialer controller")
		}

		// copied from https://github.com/XTLS/libXray/blob/main/dns/dns_android.go
		net.DefaultResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout: time.Second * 16,
				}

				dialer.Control = func(network, address string, c syscall.RawConn) error {
					return c.Control(func(fd uintptr) {
						C.libxivpn_protect(C.int(fd))
					})
				}

				log("resolver dial " + network + " " + address)

				return dialer.DialContext(ctx, network, "8.8.8.8:53")
			},
		}

	})

	log(fmt.Sprintln("start", config, socksPort, fd))

	defer func() {
		r := recover()
		if r != nil {
			b := make([]byte, 4096)
			n := runtime.Stack(b, false)
			s := string(b[:n])
			log(fmt.Sprintf("panic: %v", r))
			log(fmt.Sprintf("panic: %v", s))
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
		if logFile != nil {
			logFile.Close()
			logFile = nil
		}
	}()
	defer func() {
		r := recover()
		if r != nil {
			b := make([]byte, 4096)
			n := runtime.Stack(b, false)
			s := string(b[:n])
			log(fmt.Sprintf("panic: %v", r))
			log(fmt.Sprintf("panic: %v", s))
		}
	}()

	log("stop")

	// engine.Stop() will not close fd
	// it has no effect if engine is never started
	engine.Stop()

	if xrayServer != nil {
		err := xrayServer.Close()
		if err != nil {
			log(fmt.Sprintf("xray server close error: %v", err.Error()))
		} else {
			log("xray server closed")
		}
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
