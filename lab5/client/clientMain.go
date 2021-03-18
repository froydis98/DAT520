// A simple UDP based echo client and server.
package main

import (
	"dat520/lab5/bank"
	"dat520/lab5/multipaxos"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// OtherServer - Used to know what addrs that might be possible to connect to
type OtherServer struct {
	addr   string
	nodeID int
}

// Endpoints - Matches the endpoints in json files
type Endpoints struct {
	ID   int
	Addr string
}

// NetworkConfig - should math json files
type NetworkConfig struct {
	Endpoints []Endpoints
}

func importNetConfig(path string) (NetworkConfig, error) {
	netConfigFile, err := os.Open(path)
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
	netconf, _ := importNetConfig("clientNetConfig.json")
	netconfServers, _ := importNetConfig("netConfig.json")
	fmt.Println(netconf)
	fmt.Printf("The servers are: %v \nWrite in the index of the one you want to run: ", netconf.Endpoints)
	var serverID string
	fmt.Scanln(&serverID)
	id, err := strconv.Atoi(serverID)
	if err != nil {
		fmt.Println(err)
	}
	Servers := make([]OtherServer, 0)
	for _, endpoint := range netconfServers.Endpoints {
		Servers = append(Servers, OtherServer{endpoint.Addr, endpoint.ID})
	}
	server, _ := NewUDPServer(netconf.Endpoints[id].Addr, netconf.Endpoints[id].ID)
	clientSeq := 0
	go server.ServeUDP()
	fmt.Printf("The servers are: %v \nWrite in the index of the one you want to run: ", netconf.Endpoints)
	for {
		clientSeq++
		fmt.Println("Please enter the account number: ")
		var acc string
		fmt.Scanln(&acc)
		fmt.Println("What kind of transaction do you want to do?\n0 = Balance\n1 = Deposit\n2 = Withdrawal")
		var op bank.Operation
		fmt.Scanln(op)
		var amount int
		if op == 1 {
			fmt.Println("How much money do you want to withdraw?")
			fmt.Scanln(amount)
		} else if op == 2 {
			fmt.Println("How much money do you want to deposit?")
			fmt.Scanln(amount)
		}
		accountNr, _ := strconv.Atoi(acc)
		val := multipaxos.Value{ClientID: serverID, ClientSeq: clientSeq, Noop: true, AccountNum: accountNr, Tnx: bank.Transaction{Op: op, Amount: amount}}
		valueString, _ := json.Marshal(val)
		if acc != "" {
			for _, server := range Servers {
				_, err := SendCommand(server.addr, "ClientRequest", string(valueString))
				if err != nil {
					fmt.Println(err)
				}
			}
		}

	}
}
