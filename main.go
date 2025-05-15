package main

/*
#include <jni.h>
*/
import "C"

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

var ipcconn *os.File

// https://stackoverflow.com/questions/47644667/is-it-possible-to-use-go-to-send-and-receive-file-descriptors-over-unix-domain-s
func recv_fd() int {
	l_msg := syscall.CmsgSpace(1 * 4)
	buf := make([]byte, l_msg)
	_, _, _, _, err := syscall.Recvmsg(int(ipcconn.Fd()), nil, buf, 0)
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

func protectFd(fd int) {
	rights := syscall.UnixRights(fd)
	err := syscall.Sendmsg(int(ipcconn.Fd()), nil, rights, nil, 0)
	if err != nil {
		panic(fmt.Errorf("send to ipc conn: %w", err))
	}

	var buf [1]byte
	_, _, _, _, err = syscall.Recvmsg(int(ipcconn.Fd()), nil, buf[:], 0)
	if err != nil {
		panic(fmt.Errorf("recv from ipc conn: %w", err))
	}
}

func main() {
	unixconn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: "@xivpn_ipc", Net: "unix"})
	if err != nil {
		panic(fmt.Errorf("dial ipc conn: %w", err))
	}
	ipcconn, err = unixconn.File()
	if err != nil {
		panic(fmt.Errorf("get ipc conn file: %w", err))
	}

	fd := recv_fd()

	configBytes, err := os.ReadFile("config.json")
	if err != nil {
		panic(fmt.Errorf("read config.json: %w", err))
	}

	err = libxivpn_start(string(configBytes), 18964, fd, "", "")
	if err != nil {
		panic(fmt.Errorf("start libxivpn: %w", err))
	}
}
