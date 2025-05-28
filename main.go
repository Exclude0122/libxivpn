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

var ipcconn *net.UnixConn

func log(msg string) {
	fmt.Fprintln(os.Stderr, "libxivpn", time.Now().Format(time.DateTime), msg)
}

func ipcfd() int {
	ipcfile, _ := ipcconn.File()
	return int(ipcfile.Fd())
}

// https://stackoverflow.com/questions/47644667/is-it-possible-to-use-go-to-send-and-receive-file-descriptors-over-unix-domain-s
func recv_fd() int {
	l_msg := syscall.CmsgSpace(1 * 4)
	buf := make([]byte, l_msg)
	_, _, _, _, err := syscall.Recvmsg(ipcfd(), nil, buf, 0)
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
	if len(fds) != 1 {
		panic(fmt.Errorf("expect 1 fd, found %d", len(fds)))
	}

	return fds[0]
}

var protect_lock sync.Mutex

func protectFd(fd int) {
	rights := syscall.UnixRights(fd)

	protect_lock.Lock()

	err := syscall.Sendmsg(ipcfd(), []byte{0}, rights, nil, 0)
	if err != nil {
		panic(fmt.Errorf("send to ipc conn: %w", err))
	}

	var buf [1]byte
	_, _, _, _, err = syscall.Recvmsg(ipcfd(), buf[:], nil, 0)
	if err != nil {
		panic(fmt.Errorf("recv from ipc conn: %w", err))
	}

	protect_lock.Unlock()
}

func main() {
	config, err := os.ReadFile(filepath.Join(os.Getenv("XRAY_LOCATION_ASSET"), "config.json"))
	if err != nil {
		panic(fmt.Errorf("read config.json: %w", err))
	}

	ipcconn, err = net.DialUnix("unix", nil, &net.UnixAddr{Name: os.Getenv("IPC_PATH"), Net: "unix"})
	if err != nil {
		panic(fmt.Errorf("dial ipc conn: %w", err))
	}

	tunfd := recv_fd()

	err = libxivpn_start(string(config), 18964, tunfd)
	if err != nil {
		panic(fmt.Errorf("start libxivpn: %w", err))
	}

	fmt.Scanln()

	log("stop libxivpn")
	libxivpn_stop()
}
