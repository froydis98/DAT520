// A simple UDP based echo client and server.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type OtherServer struct {
	addr   string
	nodeID int
}

type Endpoints struct {
	Id   int
	Addr string
}

type NetworkConfig struct {
	Endpoints []Endpoints
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
	return netconf, nil
}

func main() {
	Servers := make([]OtherServer, 0)
	Servers = append(Servers, OtherServer{"127.0.0.1:5", 5})
	Servers = append(Servers, OtherServer{"127.0.0.1:1", 1})
	Servers = append(Servers, OtherServer{"127.0.0.1:2", 2})

	netconf, _ := importNetConfig()
	fmt.Printf("The servers are: %v \nWrite in the index of the one you want to run: ", netconf.Endpoints)
	var serverID string
	fmt.Scanln(&serverID)
	id, err := strconv.Atoi(serverID)
	if err != nil {
		fmt.Println(err)
	}
	server, _ := NewUDPServer(netconf.Endpoints[id].Addr, netconf.Endpoints[id].Id)

	go server.ServeUDP()
	fmt.Printf("The servers are: %v \nWrite in the index of the one you want to run: ", netconf.Endpoints)
	for {
		fmt.Println("Please enter value you want to send: ")
		first := ""
		fmt.Scanln(&first)
		if first != "" {
			for _, server := range Servers {
				_, err := SendCommand(server.addr, "ClientRequest", first)
				if err != nil {

				}
			}
		}

	}
}
