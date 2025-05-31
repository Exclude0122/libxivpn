package main

/*
#include <jni.h>
*/
import "C"

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var ipcConn *net.UnixConn

func log(msg string) {
	fmt.Fprintln(os.Stderr, "libxivpn", time.Now().Format(time.DateTime), msg)
}

func ipcFd() int {
	ipcfile, _ := ipcConn.File()
	return int(ipcfile.Fd())
}

// Receive fd from ipcConn
// https://stackoverflow.com/questions/47644667/is-it-possible-to-use-go-to-send-and-receive-file-descriptors-over-unix-domain-s
func recvFd() int {
	buffer := syscall.CmsgSpace(1 * 4)
	buf := make([]byte, buffer)
	_, _, _, _, err := syscall.Recvmsg(ipcFd(), nil, buf, 0)
	if err != nil {
		panic(fmt.Errorf("recv from ipc conn: %w", err))
	}

	msgs, err := syscall.ParseSocketControlMessage(buf)
	if err != nil {
		panic(fmt.Errorf("parse socket control msgs: %w", err))
	}
	if len(msgs) != 1 {
		panic(fmt.Errorf("expect 1 socket control msgs, found %d", len(msgs)))
	}

	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		panic("parse unix rights: " + err.Error())
	}
	if len(fds) != 1 {
		panic(fmt.Errorf("expect 1 fd, found %d", len(fds)))
	}

	return fds[0]
}

var protectLock sync.Mutex

func protectFd(fd int) {
	rights := syscall.UnixRights(fd)

	protectLock.Lock()

	err := syscall.Sendmsg(ipcFd(), []byte{0}, rights, nil, 0)
	if err != nil {
		panic(fmt.Errorf("send to ipc conn: %w", err))
	}

	var buf [1]byte
	_, _, _, _, err = syscall.Recvmsg(ipcFd(), buf[:], nil, 0)
	if err != nil {
		panic(fmt.Errorf("recv from ipc conn: %w", err))
	}

	protectLock.Unlock()
}

func main() {
	config, err := os.ReadFile(filepath.Join(os.Getenv("XRAY_LOCATION_ASSET"), "config.json"))
	if err != nil {
		panic(fmt.Errorf("read config.json: %w", err))
	}

	ipcConn, err = net.DialUnix("unix", nil, &net.UnixAddr{Name: os.Getenv("IPC_PATH"), Net: "unix"})
	if err != nil {
		panic(fmt.Errorf("dial ipc conn: %w", err))
	}

	tunFd := recvFd()

	err = libxivpn_start(string(config), 18964, tunFd)
	if err != nil {
		panic(fmt.Errorf("start libxivpn: %w", err))
	}

	fmt.Scanln()

	log("stop libxivpn")
	libxivpn_stop()
}
