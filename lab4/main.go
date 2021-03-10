// A simple UDP based echo client and server.
package main

import (
	"dat520/lab3/failuredetector"
	"dat520/lab3/leaderdetector"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type OtherServer struct {
	addr   string
	nodeID int
}

type Endpoints struct {
	Id int
	Addr string
}

type NetworkConfig struct {
	Endpoints []Endpoints
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

func importNetConfig() (NetworkConfig, error) {
	netConfigFile, err := os.Open("netConfig.json")
	if err != nil {
		fmt.Println(err)
	}
	defer netConfigFile.Close()
	byteValue, _ := ioutil.ReadAll(netConfigFile)
	var netconf NetworkConfig
	err = json.Unmarshal(byteValue, &netconf)
	if err != nil {
		fmt.Println(err)
		return netconf, err
	}
	fmt.Println(netconf)
	return netconf, nil
}

func main() {
	nodeIDs := make([]int, 0)
	OtherServers := make([]OtherServer, 0)
	netconf, _ := importNetConfig()
	fmt.Println("Write in ID")
	var serverID string
	fmt.Scanln(&serverID)
	id, err := strconv.Atoi(serverID)
	if err != nil {
		fmt.Println(err)
	}
	server, _ := NewUDPServer(netconf.Endpoints[id].Addr, netconf.Endpoints[id].Id)

	go server.ServeUDP()

	for _, endpoints := range netconf.Endpoints {
		OtherServers = append(OtherServers, OtherServer{endpoints.Addr, endpoints.Id})
		nodeIDs = append(nodeIDs, endpoints.Id)
	}

	fmt.Println("outside of for loop", nodeIDs)
	nld := leaderdetector.NewMonLeaderDetector(nodeIDs)
	delta := time.Second * 5
	hbSend := make(chan failuredetector.Heartbeat, 4294967)
	nfd := failuredetector.NewEvtFailureDetector(server.id, nodeIDs, nld, delta, hbSend)
	nfd.Start()
	for {
		fmt.Printf("\n\nThe leader is: %v\n", nld.Leader())
		fmt.Printf("The suspected nodes are: %v\n", nld.Suspected)
		select {
		case <- hbSend:
			for _, server := range OtherServers {
				time.Sleep(time.Millisecond * 600)
				HeartbeatString := strconv.Itoa(nfd.ID) + "," + strconv.Itoa(server.nodeID) + "," + "true"
				res, err := SendCommand(server.addr, "HeartBeat", HeartbeatString)
			if err != nil {
				fmt.Println(err)
			}

			splittedRes := strings.Split(res, ",")
			if len(splittedRes) > 2 {
				state := false
				if splittedRes[2] == "true" {
					state = true
				}
				from, _ := strconv.Atoi(splittedRes[0])
				to, _ := strconv.Atoi(splittedRes[1])
				nfd.DeliverHeartbeat(failuredetector.Heartbeat{From: from, To: to, Request: state})
			}
			}
		}
	}
}
