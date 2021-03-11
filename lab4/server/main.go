// A simple UDP based echo client and server.
package main

import (
	"dat520/lab3/failuredetector"
	"dat520/lab3/leaderdetector"
	"dat520/lab4/multipaxos"
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
	Id   int
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
	return netconf, nil
}

var ClientChan chan string
var PrepareIn chan string
var PromiseIn chan string

func main() {
	nodeIDs := make([]int, 0)
	OtherServers := make([]OtherServer, 0)
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

	for _, endpoints := range netconf.Endpoints {
		OtherServers = append(OtherServers, OtherServer{endpoints.Addr, endpoints.Id})
		nodeIDs = append(nodeIDs, endpoints.Id)
	}
	nld := leaderdetector.NewMonLeaderDetector(nodeIDs)
	delta := time.Second * 2
	hbSend := make(chan failuredetector.Heartbeat, 4294967)
	nfd := failuredetector.NewEvtFailureDetector(server.id, nodeIDs, nld, delta, hbSend)
	prepareOut := make(chan multipaxos.Prepare, 4294967)
	acceptOut := make(chan multipaxos.Accept, 4294967)
	promiseOut := make(chan multipaxos.Promise, 4294967)
	PrepareIn = make(chan string, 4294967)
	PromiseIn = make(chan string, 4294967)
	LearnOut := make(chan multipaxos.Learn, 4294967)

	var proposer *multipaxos.Proposer
	var acceptor *multipaxos.Acceptor
	// var learner *multipaxos.Learner
	proposer = multipaxos.NewProposer(server.id, len(OtherServers), -1, nld, prepareOut, acceptOut)
	acceptor = multipaxos.NewAcceptor(server.id, promiseOut, LearnOut)
	proposer.Start()
	acceptor.Start()

	ClientChan = make(chan string, 10000)
	// var decidedValue *multipaxos.DecidedValue

	nfd.Start()
	for {
		fmt.Printf("\n\nThe leader is: %v\n", nld.Leader())
		fmt.Printf("The suspected nodes are: %v\n", nld.Suspected)
		select {
		case <-hbSend:
			for _, server := range OtherServers {
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
		case input := <-ClientChan:
			splitInput := strings.Split(input, ",")
			res, _ := strconv.Atoi(splitInput[1])
			b1, _ := strconv.ParseBool(splitInput[2])
			value := multipaxos.Value{splitInput[0], res, b1, splitInput[3]}
			fmt.Println(value)

		case pr := <-prepareOut:
			fmt.Println(pr.Crnd, "Crnd!!!!!!!!!!!!! Prepare out from proposer")
			if server.id == nld.CurrentLeader {
				for _, Otherserver := range OtherServers {
					prString, _ := json.Marshal(pr)
					res, err := SendCommand(Otherserver.addr, "Prepare", string(prString))
					fmt.Println(res, err)

				}
			}

		case pr := <-PrepareIn:

			var prepare = &multipaxos.Prepare{}
			err := json.Unmarshal([]byte(pr), prepare)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Prepare After sendt", *&prepare.Crnd, "assssssssssssssssssssssssss")
			acceptor.DeliverPrepare(*prepare)

		case promise := <-promiseOut:
			fmt.Println(promise.Rnd, "Crnd!!!!!!!!!!!!! Promise out from Accepter")

			for _, server := range OtherServers {
				if server.nodeID == promise.To {
					promiseString, _ := json.Marshal(promise)
					SendCommand(server.addr, "Promise", string(promiseString))
				}
			}
		case promiseStringIn := <-PromiseIn:
			var promise = &multipaxos.Promise{}
			err := json.Unmarshal([]byte(promiseStringIn), promise)
			fmt.Println(promise.Rnd, "Crnd!!!!!!!!!!!!! Promise in to proposer")

			if err != nil {
				fmt.Println(err)
			}
			proposer.DeliverPromise(*promise)
			// case ac := <-acceptChan:
			// 	fmt.Printf("\nGot an accept: %v", ac)
			// 	acceptor.DeliverAccept(multipaxos.Accept{From: ac.From, Slot: ac.Slot, Rnd: ac.Rnd, Val: ac.Val})
			// case pr := <-promiseChan:
			// 	fmt.Printf("\nGot a promise: %v", pr)
			// 	proposer.DeliverPromise(multipaxos.Promise{To: pr.To, From: pr.From, Rnd: pr.Rnd, Slots: pr.Slots})
			// case lr := <-learnChan:
			// 	fmt.Printf("\nGot a learn: %v", lr)
			// 	learner.DeliverLearn(multipaxos.Learn{From: lr.From, Slot: lr.Slot, Rnd: lr.Rnd, Val: lr.Val})
		}
	}
}
