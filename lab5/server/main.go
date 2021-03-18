// A simple UDP based echo client and server.
package main

import (
	"dat520/lab3/failuredetector"
	"dat520/lab3/leaderdetector"
	"dat520/lab5/bank"
	"dat520/lab5/multipaxos"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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

// ClientChan - global channel
var ClientChan chan string

// PrepareIn - global channel
var PrepareIn chan string

// PromiseIn - global channel
var PromiseIn chan string

// AcceptIn - global channel
var AcceptIn chan string

// LearnIn - global channel
var LearnIn chan string

// UpdateAdu - global channel
var UpdateAdu chan string

// ReConfig - global channel
var ReConfig chan string

func main() {
	nodeIDs := make([]int, 0)
	clientAddr := make([]string, 0)
	OtherServers := make([]OtherServer, 0)
	currentServers := make(map[int]OtherServer, 0)

	decidedValueBuffer := make(map[int]multipaxos.DecidedValue, 0)
	netconf, _ := importNetConfig("./netConfig.json")
	netConfClients, _ := importNetConfig("./clientNetConfig.json")
	fmt.Printf("The servers are: %v \n", netconf.Endpoints)
	addrs, _ := net.InterfaceAddrs()
	var addr string
	for _, i := range addrs {
		if strings.Contains(i.String(), "192") {
			addr = i.String()
		}
	}
	splittedAddr := strings.Split(addr, "/")
	splitAgain := strings.Split(splittedAddr[0], ".")
	id, _ := strconv.Atoi(splitAgain[3])
	id = id - 2
	for _, endpoint := range netconf.Endpoints {
		OtherServers = append(OtherServers, OtherServer{endpoint.Addr, endpoint.ID})
		nodeIDs = append(nodeIDs, endpoint.ID)
	}
	for _, endpoint := range netConfClients.Endpoints {
		clientAddr = append(clientAddr, endpoint.Addr)
	}
	server, _ := NewUDPServer(netconf.Endpoints[id].Addr, netconf.Endpoints[id].ID)

	go server.ServeUDP()

	nld := leaderdetector.NewMonLeaderDetector(nodeIDs)
	delta := time.Second * 3
	hbSend := make(chan failuredetector.Heartbeat, 4294967)
	nfd := failuredetector.NewEvtFailureDetector(server.id, nodeIDs, nld, delta, hbSend)
	prepareOut := make(chan multipaxos.Prepare, 4294967)
	acceptOut := make(chan multipaxos.Accept, 4294967)
	promiseOut := make(chan multipaxos.Promise, 4294967)
	PrepareIn = make(chan string, 4294967)
	PromiseIn = make(chan string, 4294967)
	AcceptIn = make(chan string, 4294967)
	decidedOut := make(chan multipaxos.DecidedValue, 4294967)
	learnOut := make(chan multipaxos.Learn, 4294967)
	LearnIn = make(chan string, 4294967)
	ClientChan = make(chan string, 4294967)
	UpdateAdu = make(chan string, 4294967)
	ReConfig = make(chan string, 4294967)
	for {
		time.Sleep(time.Second)
		var NmrServers int
		fmt.Println("We are waiting in first For loop for current servers")
		select {
		case nmrServers := <-ReConfig:
			currentServers := make(map[int]OtherServer, 0)
			NmrServers, _ = strconv.Atoi(nmrServers)
			for i := 0; i < NmrServers; i++ {
				{
					currentServers[OtherServers[i].nodeID] = OtherServers[i]
					OtherServers = append(OtherServers, OtherServers[i])
					OtherServers = OtherServers[1:]
				}
			}
		default:
			continue
		}
		_, ok := currentServers[id]
		if ok == true {
			var proposer *multipaxos.Proposer
			var acceptor *multipaxos.Acceptor
			var learner *multipaxos.Learner
			bankAccounts := make(map[int]bank.Account)
			adu := -1
			proposer = multipaxos.NewProposer(server.id, len(OtherServers), adu, nld, prepareOut, acceptOut)
			acceptor = multipaxos.NewAcceptor(server.id, promiseOut, learnOut)
			learner = multipaxos.NewLearner(server.id, len(OtherServers), decidedOut)
			proposer.Start()
			acceptor.Start()
			learner.Start()
			nfd.Start()

			for {
				fmt.Printf("\n\nThe leader is: %v\n", nld.Leader())
				fmt.Printf("The suspected nodes are: %v\n", nld.Suspected)
				select {
				case <-hbSend:
					for _, server := range OtherServers {
						fmt.Println("these are the servers", server)
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
							fmt.Println("Delivering a heartbeart From:", from, " to ", to)
							nfd.DeliverHeartbeat(failuredetector.Heartbeat{From: from, To: to, Request: state})
						}
					}
				case pr := <-prepareOut:
					fmt.Println("The current round is: ", pr.Crnd, "\nPrepare out from proposer")
					for _, Otherserver := range OtherServers {
						prString, _ := json.Marshal(pr)
						res, err := SendCommand(Otherserver.addr, "Prepare", string(prString))
						fmt.Println(res, err)

					}

				case pr := <-PrepareIn:
					var prepare = &multipaxos.Prepare{}
					err := json.Unmarshal([]byte(pr), prepare)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("Prepare After sendt \nThe round is: ", *&prepare.Crnd)
					acceptor.DeliverPrepare(*prepare)

				case promise := <-promiseOut:
					fmt.Println("The round is: ", promise.Rnd, "\nPromise out from Accepter")

					for _, server := range OtherServers {
						if server.nodeID == promise.To {
							promiseString, _ := json.Marshal(promise)
							SendCommand(server.addr, "Promise", string(promiseString))
						}
					}
				case promiseStringIn := <-PromiseIn:
					var promise = &multipaxos.Promise{}
					err := json.Unmarshal([]byte(promiseStringIn), promise)
					fmt.Println("The round is: ", promise.Rnd, "\nPromise in to proposer")

					if err != nil {
						fmt.Println(err)
					}
					proposer.DeliverPromise(*promise)
				case input := <-ClientChan:
					fmt.Println("Inside the client channel")
					clientVal := &multipaxos.Value{}
					err := json.Unmarshal([]byte(input), clientVal)
					if err != nil {
						fmt.Println(err)
					}
					value := multipaxos.Value{ClientID: clientVal.ClientID, ClientSeq: clientVal.ClientSeq, Noop: clientVal.Noop, AccountNum: clientVal.AccountNum, Tnx: clientVal.Tnx}
					fmt.Println(value)
					proposer.DeliverClientValue(value)

				case acceptOut := <-acceptOut:
					fmt.Println("Inside Accept Out from Proposer")
					acceptOutString, _ := json.Marshal(acceptOut)
					for _, otherserver := range OtherServers {
						SendCommand(otherserver.addr, "AcceptIn", string(acceptOutString))

					}
				case acceptIn := <-AcceptIn:
					fmt.Println("Inside acceptIn, right before Acceptor")
					var acceptmsg = &multipaxos.Accept{}
					err := json.Unmarshal([]byte(acceptIn), acceptmsg)
					if err != nil {
						fmt.Println(err)
					}
					acceptor.DeliverAccept(*acceptmsg)
				case learnmsg := <-learnOut:
					fmt.Println("Insiden Learn Out from Acceptor")
					learnString, _ := json.Marshal(learnmsg)
					for _, otherserver := range OtherServers {
						SendCommand(otherserver.addr, "Learn", string(learnString))
					}
				case learninmsg := <-LearnIn:
					var learnmsg = &multipaxos.Learn{}
					err := json.Unmarshal([]byte(learninmsg), learnmsg)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("The learn message is: ", *learnmsg)
					learner.DeliverLearn(*learnmsg)
				case decidedout := <-decidedOut:
					if int(decidedout.SlotID) > (adu + 1) {
						decidedValueBuffer[int(decidedout.SlotID)] = decidedout
						return
					}
					if decidedout.Value.Noop != true {
						_, ok := bankAccounts[decidedout.Value.AccountNum]
						if ok == false {
							fmt.Println("Bank account not found, creating new account")
							bankAccounts[decidedout.Value.AccountNum] = bank.Account{Number: decidedout.Value.AccountNum, Balance: 0}
						}
						for _, account := range bankAccounts {
							fmt.Println("THis is account", account)
						}
						account := bankAccounts[decidedout.Value.AccountNum]
						transaction := account.Process(decidedout.Value.Tnx)
						bankAccounts[decidedout.Value.AccountNum] = account
						fmt.Println(transaction)
						if server.id == nld.CurrentLeader {
							for _, client := range clientAddr {
								out := multipaxos.Response{ClientID: decidedout.Value.ClientID, ClientSeq: decidedout.Value.ClientSeq, TnxRes: transaction}
								SendCommand(client, "NewValue", out.String())
							}
						}
					}
					if server.id == nld.CurrentLeader {
						for _, otherServer := range OtherServers {
							SendCommand(otherServer.addr, "UpdateAdu", "test")
						}
					}
					if _, ok := decidedValueBuffer[adu+1]; ok {
						value := decidedValueBuffer[adu+1]
						decidedOut <- value
					}

				case <-UpdateAdu:
					adu++
					proposer.IncrementAllDecidedUpTo()
				}
			}
		}
		fmt.Println("End of for loop")
	}
}
