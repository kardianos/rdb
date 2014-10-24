// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package ssrp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	maxPacketSize = 1024 * 64
	writeMilli    = 30
	readMilli     = 60
)

func FetchInstanceInfo(server, instance string) (*InstanceInfo, error) {
	list, err := fetch(server, instance, false)
	if err != nil {
		return nil, err
	}
	return &list[0], nil
}

func FetchInstanceInfoList(server string) ([]InstanceInfo, error) {
	return fetch(server, "", true)
}

func fetch(server, instance string, all bool) ([]InstanceInfo, error) {
	serverAddress, err := net.ResolveUDPAddr("udp", server+":1434")
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	err = conn.SetReadBuffer(maxPacketSize * 3)
	if err != nil {
		return nil, err
	}

	var msg []byte
	if all {
		msg = []byte{3}
	} else {
		msg = make([]byte, 34)
		msg[0] = 4
		nameLen := copy(msg[1:], instance)
		msg[nameLen+2] = 0
	}

	recvPacket := make([]byte, maxPacketSize) // Max size.

	var info []InstanceInfo

	for i := 0; i < 3; i++ {
		err = conn.SetWriteDeadline(time.Now().Add(writeMilli * time.Millisecond))
		if err != nil {
			return nil, err
		}
		_, err := conn.Write(msg)
		if err != nil {
			return nil, err
		}
		info, err = readResult(conn, recvPacket)
		if err != nil {
			return nil, err
		}
		if info != nil {
			break
		}
		time.Sleep(time.Millisecond * 300)
	}
	if info == nil {
		return nil, fmt.Errorf("Unable to fetch server connection information.")
	}

	return info, nil
}

type InstanceInfo struct {
	IP        net.IP
	Server    string
	Instance  string
	Tcp       int
	NamedPipe string
}

func readResult(conn *net.UDPConn, bb []byte) ([]InstanceInfo, error) {
	err := conn.SetReadDeadline(time.Now().Add(readMilli * time.Millisecond))
	if err != nil {
		return nil, err
	}
	n, addr, err := conn.ReadFromUDP(bb)
	if err != nil {
		return nil, err
	}
	if n < 3 {
		return nil, nil
	}
	if bb[0] != 5 {
		return nil, nil
	}
	msgLen := int(binary.LittleEndian.Uint16(bb[1:]))
	if msgLen+3 != n {
		return nil, nil
	}

	msg := bb[3 : 3+msgLen]
	msgList := bytes.Split(msg, []byte(";"))

	infoList := make([]InstanceInfo, 1)
	infoListIndex := 0
	infoList[infoListIndex].IP = addr.IP

loop:
	for i := 0; i < len(msgList); i += 2 {
		key := string(msgList[i])
		value := string(msgList[i+1])
		switch key {
		case "":
			if value == "" {
				break loop
			}
			infoList = append(infoList, InstanceInfo{IP: addr.IP})
			infoListIndex++
			i--
		case "ServerName":
			infoList[infoListIndex].Server = value
		case "InstanceName":
			infoList[infoListIndex].Instance = value
		case "tcp":
			infoList[infoListIndex].Tcp, err = strconv.Atoi(value)
		case "np":
			infoList[infoListIndex].NamedPipe = value
		}
	}
	return infoList, nil
}
