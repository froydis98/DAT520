// A simple UDP based echo client and server.
package main

import (
	"flag"
	"fmt"
	// "net"
	"os"
	"strconv"
	"strings"
	"time"
)

type OtherServer struct {
	addr   string
	nodeID int
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	nodeIDs := make([]int, 0)
	OtherServers := make([]OtherServer, 0)
	endpoint1 := "localhost:12111"
	endpoint2 := "localhost:12112"
	endpoint3 := "localhost:12113"

	endpoints := make([]string, 0)
	endpoints = append(endpoints, endpoint1)
	endpoints = append(endpoints, endpoint2)
	endpoints = append(endpoints, endpoint3)

	server1, _ := NewUDPServer(endpoint3, 5)

	// conn := make([]*net.UDPAddr, 0)
	connstrings := make([]string, 0)

	go server1.ServeUDP()

	// waiting := true
	// for waiting {
	// 	for _, endpoint := range endpoints {
	// 		res, err := SendCommand(endpoint, "AddServerIp", "THis is a test")
	// 		fmt.Println(res)
	// 		check(err)
	// 		if res != "" && !stringInSlice(res, connstrings) {
	// 			s := strings.Split(res, ",")
	// 			fmt.Println(s)
	// 			udpAddr, _ := net.ResolveUDPAddr("udp", s[0])
	// 			nodeID, _ := strconv.Atoi(s[1])
	// 			if intInSlice(nodeID, nodeIDs) {
	// 				waiting = false
	// 			}
	// 			anotherServer := OtherServer{s[0], nodeID}
	// 			OtherServers = append(OtherServers, anotherServer)
	// 			nodeIDs = append(nodeIDs, nodeID)
	// 			connstrings = append(connstrings, res)
	// 			conn = append(conn, udpAddr)
	// 		}
	// 		if len(conn) > 2 {
	// 			waiting = false
	// 		}
	// 	}
	// }
	
	OtherServers = append(OtherServers, OtherServer{endpoint1, 10})
	OtherServers = append(OtherServers, OtherServer{endpoint2, 20})
	OtherServers = append(OtherServers, OtherServer{endpoint3, 5})
	nodeIDs = append(nodeIDs, 10)
	nodeIDs = append(nodeIDs, 20)
	nodeIDs = append(nodeIDs, 5)



	fmt.Println("outside of for loop", connstrings)
	fmt.Println("outside of for loop", nodeIDs)
	nld := NewMonLeaderDetector(nodeIDs)
	NewSuspecter := nld
	delta := time.Second * 5
	hbSend := make(chan Heartbeat, 16)
	nfd := NewEvtFailureDetector(server1.id, nodeIDs, NewSuspecter, delta, hbSend)
	nfd.Start()
	for {
		fmt.Println("")
		fmt.Println("Current leader is: ", nld.Leader())

		if nfd.suspected[nld.Leader()] {
			nld.ChangeLeader()
		}
		fmt.Println("Suspected Nodes are: ", nfd.suspected)

		for _, server := range nfd.nodeIDs {
			time.Sleep(time.Millisecond * 600)

			tempAddr := ""
			for _, otherservers := range OtherServers {

				if server == otherservers.nodeID {
					tempAddr = otherservers.addr
				}
			}

			if tempAddr != "" {
				res := strconv.Itoa(nfd.id)
				server := strconv.Itoa(server)
				HeartbeatString := res + "," + server + "," + "true"
				res, err := SendCommand(tempAddr, "HeartBeat", HeartbeatString)
				fmt.Println("Response from other servers in format from-to: ", res)
				if err != nil {
					fmt.Println(err)
				}

				splittedRes := strings.Split(res, ",")
				if len(splittedRes) > 2 {
					state := false

					if splittedRes[2] == "true" {
						state = true
					}

					var heartBeat Heartbeat
					from, _ := strconv.Atoi(splittedRes[0])
					to, _ := strconv.Atoi(splittedRes[1])
					heartBeat.From = from
					if nfd.suspected[heartBeat.From] {
						nld.Restore(heartBeat.From)
						nfd.suspected[heartBeat.From] = false
					}
					heartBeat.To = to
					heartBeat.Request = state
					nfd.DeliverHeartbeat(heartBeat)
				}
			}
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
