// A simple UDP based echo client and server.
package main

import (
	"bufio"
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
	for {
		clientSeq++
		accountNr, OP, amount := AskForInfo()
		val := multipaxos.Value{ClientID: serverID, ClientSeq: clientSeq, Noop: true, AccountNum: accountNr, Tnx: bank.Transaction{Op: bank.Operation(OP), Amount: amount}}
		valueString, _ := json.Marshal(val)
		for _, server := range Servers {
			_, err := SendCommand(server.addr, "ClientRequest", string(valueString))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}


func AskForInfo() (int, int, int) {
	fmt.Println("Please enter the account number: ")
	var acc = bufio.NewScanner(os.Stdin)
	acc.Scan()
	if acc.Text() != "" {
		fmt.Println("What kind of transaction do you want to do?\n0 = Balance\n1 = Deposit\n2 = Withdrawal")
		var op = bufio.NewScanner(os.Stdin)
		op.Scan()
		OP, _ := strconv.Atoi(op.Text())
		if op.Text() != "" {
			var am = bufio.NewScanner(os.Stdin)
			if OP == 1 {
				fmt.Println("How much money do you want to withdraw?")
				am.Scan()
			} else if OP == 2 {
				fmt.Println("How much money do you want to deposit?")
				am.Scan()
			} else {
				fmt.Println("You want to see the balance")
			}
			if am.Text() != "" {
				accountNr, _ := strconv.Atoi(acc.Text())
				amount, _ := strconv.Atoi(am.Text())
				return accountNr, OP, amount
			}
		}
	}
	return -1, -1, 0
}