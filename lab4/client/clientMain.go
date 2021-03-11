// A simple UDP based echo client and server.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	// var proposer *multipaxos.Proposer
	// var prepareChan chan multipaxos.Prepare
	// var acceptChan chan multipaxos.Accept
	// var acceptor *multipaxos.Acceptor
	// var promiseChan chan multipaxos.Promise
	// var learnChan chan multipaxos.Learn
	// var learner *multipaxos.Learner
	// var decidedValue *multipaxos.DecidedValue

	for {
		fmt.Println("Please enter value you want to send: ")
		var first string
		fmt.Scanln(&first)
		for _, server := range Servers {
			res, err := SendCommand(server.addr, "ClientRequest", first)
			if err != nil {
				fmt.Println(err)

			}
			fmt.Println(res, "dette er res")
		}

	}
}
