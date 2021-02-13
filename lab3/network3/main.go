// A simple UDP based echo client and server.
package main

import (
	"flag"
	"fmt"
	"net"
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
	// De andre serverne må jeg lage i andre filer
	// server2, _ := NewUDPServer(endpoint2, 2)
	// server3, _ := NewUDPServer(endpoint3, 3)

	conn := make([]*net.UDPAddr, 0)
	connstrings := make([]string, 0)

	go server1.ServeUDP()
	// De andre serverne må jeg lage i andre filer
	// go server2.ServeUDP()
	// go server3.ServeUDP()
	waiting := true
	for waiting {
		for _, endpoint := range endpoints {
			res, err := SendCommand(endpoint, "AddServerIp", "THis is a test")
			fmt.Println(res)
			check(err)
			if res != "" && !stringInSlice(res, connstrings) {
				s := strings.Split(res, ",")
				fmt.Println(s)
				udpAddr, _ := net.ResolveUDPAddr("udp", s[0])
				nodeID, _ := strconv.Atoi(s[1])
				anotherServer := OtherServer{s[0], nodeID}
				OtherServers = append(OtherServers, anotherServer)
				nodeIDs = append(nodeIDs, nodeID)
				connstrings = append(connstrings, res)
				conn = append(conn, udpAddr)
			}
			if len(conn) > 2 {
				waiting = false
			}
		}
	}

	// for _,conns:= range connstrings{
	// 	res,_ := SendCommand(conns, "Servers", "Here we have a list of all servers")
	// 	fmt.Println(res)
	// }

	fmt.Println("outside of for loop", connstrings)
	fmt.Println("outside of for loop", nodeIDs)
	nld := NewMonLeaderDetector(nodeIDs)
	NewSuspecter := nld
	delta := time.Second * 10
	hbSend := make(chan Heartbeat, 16)
	nfd := NewEvtFailureDetector(server1.id, nodeIDs, NewSuspecter, delta, hbSend)
	nfd.Start()
	for {

		fmt.Println("")
		fmt.Println("Current leader is: ", nld.Leader())

		if nfd.suspected[nld.Leader()] {
			nld.ChangeLeader()
		}

		for _, server := range nfd.nodeIDs {
			time.Sleep(time.Second)
			fmt.Println("Suspected Nodes are: ", nfd.suspected)

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
		//	nfd.DeliverHeartbeat()
		//  kan sette setdealine for write og read, aka hvis de ikke finner noe går de videre
		//  se her for mer : https://golang.org/pkg/net/
		//  bruk ctrl f setdeadline for å finne mer om dette
		//  her må jeg lahe leaderdetector og failure detector koden
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
