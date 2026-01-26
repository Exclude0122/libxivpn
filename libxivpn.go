package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"time"

	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all"
	"github.com/xtls/xray-core/transport/internet"
)

var xrayServer *core.Instance
var registerControllerOnce sync.Once

func libxivpn_version() string {
	return strings.Join(core.VersionStatement(), "\n")
}

func libxivpn_start(config string, fd_ int) error {
	// register socket controller
	registerControllerOnce.Do(func() {
		log("register controller once")

		err := internet.RegisterDialerController(func(network, address string, conn syscall.RawConn) error {
			return conn.Control(protectFd)
		})
		if err != nil {
			log("failed to register dialer controller")
		}

		err = internet.RegisterListenerController(func(network, address string, conn syscall.RawConn) error {
			return conn.Control(protectFd)
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
					return c.Control(protectFd)
				}

				log("resolver dial " + network + " " + address)

				return dialer.DialContext(ctx, network, "8.8.8.8:53")
			},
		}
	})

	log(fmt.Sprintln("start", config, fd_))

	os.Setenv("XRAY_TUN_FD", strconv.Itoa(fd_))

	// xray

	xrayConfig, err := core.LoadConfig("json", strings.NewReader(config))
	if err != nil {
		log("libxivpn_start xray1: " + err.Error())
		return err
	}

	xrayServer, err = core.New(xrayConfig)
	if err != nil {
		log("libxivpn_start xray2: " + err.Error())
		return err
	}

	err = xrayServer.Start()
	if err != nil {
		log("libxivpn_start xray3: " + err.Error())
		return err
	}

	return nil
}

func libxivpn_stop() {
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
