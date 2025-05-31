package main

/*
#include <jni.h>
*/
import "C"

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
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

var ipcWriteLock sync.Mutex
var protectLock sync.Mutex
var protectDone = make(chan int)

func protectFd(fd int) {
	ipcWriteLock.Lock()
	defer ipcWriteLock.Unlock()

	protectLock.Lock()
	defer protectLock.Unlock()

	log(fmt.Sprintf("protectFd %d start", fd))

	rights := syscall.UnixRights(fd)

	err := syscall.Sendmsg(ipcFd(), []byte("protect\n"), rights, nil, 0)
	if err != nil {
		panic(fmt.Errorf("send to ipc conn: %w", err))
	}

	log(fmt.Sprintf("protectFd %d waiting ack", fd))

	<-protectDone // wait for protect to finish

	log(fmt.Sprintf("protectFd %d end", fd))

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

	log("ipc connected")

	tunFd := recvFd()

	log(fmt.Sprintf("tun fd %d", tunFd))

	go func() {
		log("xray starting")
		err = libxivpn_start(string(config), 18964, tunFd)
		if err != nil {
			panic(fmt.Errorf("start libxivpn: %w", err))
		}
		log("xray started")
	}()

	// hancle ipc packets

	scanner := bufio.NewScanner(ipcConn)

ipcLoop:
	for {
		log("ipc loop")

		if !scanner.Scan() {
			if scanner.Err() != nil {
				panic("ipc scan: " + scanner.Err().Error())
			}
			break
		}

		line := scanner.Text()

		splits := strings.Split(line, " ")

		log(fmt.Sprintf("ipc packet: %v", splits))

		switch splits[0] {

		case "ping":

			ipcWriteLock.Lock()
			_, err := ipcConn.Write([]byte("pong\n"))
			if err != nil {
				panic("ipc write: " + err.Error())
			}
			ipcWriteLock.Unlock()

		case "pong":
			// ignored

		case "protect_ack":
			protectDone <- 1

		case "stop":
			break ipcLoop

		}
	}

	log("stop libxivpn")
	libxivpn_stop()
}
