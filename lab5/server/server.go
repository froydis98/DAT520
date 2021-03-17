// Leave an empty line above this comment.
package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// UDPServer implements the UDP Echo Server specification found at
// https://github.com/COURSE_TAG/assignments/tree/master/lab2/README.md#udp-echo-server
type UDPServer struct {
	conn *net.UDPConn
	adr  *net.UDPAddr
	id   int
}

// NewUDPServer returns a new UDPServer listening on addr. It should return an
// error if there was any problem resolving or listening on the provided addr.
func NewUDPServer(addr string, id int) (*UDPServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &UDPServer{udpConn, udpAddr, id}, nil
}

// ServeUDP starts the UDP server's read loop. The server should read from its
// listening socket and handle incoming client requests as according to the
// the specification.
func (u *UDPServer) ServeUDP() {
	const maxBufferSize = 400000

	buffer := make([]byte, maxBufferSize)
	for {
		n, a, _ := u.conn.ReadFromUDP(buffer)
		stringen := string(buffer[:n])
		s := strings.Split(stringen, "|:|")
		if len(s) > 2 {
			fmt.Println(s[0])
			byttes := []byte("Unknown command")
			u.conn.WriteTo(byttes, a)
		} else {
			var newString string
			switch s[0] {
			case "AddServerIp":
				newstringId := strconv.Itoa(u.id)
				newString = u.adr.String() + "," + newstringId
			case "HeartBeat":
				splitter := strings.Split(s[1], ",")
				if splitter[2] == "false" {
					newString = splitter[1] + "," + splitter[0] + "," + "true"
				} else {
					newString = splitter[1] + "," + splitter[0] + "," + "false"
				}
			case "ClientRequest":
				ClientChan <- s[1]
			case "Prepare":
				PrepareIn <- s[1]
			case "Promise":
				PromiseIn <- s[1]
			case "AcceptIn":
				AcceptIn <- s[1]
			case "Learn":
				LearnIn <- s[1]
			case "UpdateAdu":
				UpdateAdu <- s[1]
			default:
				newString = "Unknown command"
			}
			byttes := []byte(newString)
			u.conn.WriteTo(byttes, a)
		}
	}
}

// socketIsClosed is a helper method to check if a listening socket has been
// closed.
func socketIsClosed(err error) bool {
	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}
	return false
}
func SendCommand(udpAddr, cmd, txt string) (string, error) {
	addr, err := net.ResolveUDPAddr("udp", udpAddr)

	if err != nil {
		return "", err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	var buf [512]byte
	cmdTxt := fmt.Sprintf("%v|:|%v", cmd, txt)
	_, err = conn.Write([]byte(cmdTxt))

	if err != nil {
		return "", err
	}
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 300))
	n, err := conn.Read(buf[0:])

	if err != nil {

		return "", nil
	}

	return string(buf[0:n]), nil
}
