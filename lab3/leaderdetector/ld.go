package leaderdetector

import (
	"sort"
)

// A MonLeaderDetector represents a Monarchical Eventual Leader Detector as
// described at page 53 in:
// Christian Cachin, Rachid Guerraoui, and Luís Rodrigues: "Introduction to
// Reliable and Secure Distributed Programming" Springer, 2nd edition, 2011.
type MonLeaderDetector struct {
	CurrentLeader  int
	Suspected      map[int]bool // map of node ids  considered suspected
	NodeIDs        []int
	LeaderChannels []chan<- int
}

// NewMonLeaderDetector returns a new Monarchical Eventual Leader Detector
// given a list of node ids.
func NewMonLeaderDetector(nodeIDs []int) *MonLeaderDetector {
	m := &MonLeaderDetector{}
	suspected := make(map[int]bool)
	m.Suspected = suspected
	sort.Ints(nodeIDs)
	if nodeIDs[len(nodeIDs)-1] < 0 {
		m.CurrentLeader = UnknownID
	} else {
		m.CurrentLeader = nodeIDs[len(nodeIDs)-1]
	}
	m.NodeIDs = nodeIDs
	return m
}

// Leader returns the current leader. Leader will return UnknownID if all nodes
// are suspected.
func (m *MonLeaderDetector) Leader() int {
	if m.CurrentLeader == UnknownID {
		return UnknownID
	}
	return m.CurrentLeader
}

// Suspect instructs the leader detector to consider the node with matching
// id as suspected. If the suspect indication result in a leader change
// the leader detector should publish this change to its subscribers.
func (m *MonLeaderDetector) Suspect(id int) {
	m.Suspected[id] = true
	m.ChangeLeader()
}

// Restore instructs the leader detector to consider the node with matching
// id as restored. If the restore indication result in a leader change
// the leader detector should publish this change to its subscribers.
func (m *MonLeaderDetector) Restore(id int) {
	m.Suspected[id] = false
	m.ChangeLeader()
}

// Subscribe returns a buffered channel which will be used by the leader
// detector to publish the id of the highest ranking node.
// The leader detector will publish UnknownID if all nodes become suspected.
// Subscribe will drop publications to slow subscribers.
// Note: Subscribe returns a unique channel to every subscriber;
// it is not meant to be shared.
func (m *MonLeaderDetector) Subscribe() <-chan int {
	leaderChan := make(chan int, 8)
	m.LeaderChannels = append(m.LeaderChannels, leaderChan)
	return leaderChan
}

func (m *MonLeaderDetector) ChangeLeader() {
	allTrue := false
	tempLeader := m.CurrentLeader
	for _, node := range m.NodeIDs {
		if !m.Suspected[node] {
			if node < 0 {
				m.CurrentLeader = UnknownID
			} else {
				m.CurrentLeader = node
			}
			allTrue = true
		}
	}
	if !allTrue {
		m.CurrentLeader = UnknownID
	}
	if tempLeader != m.CurrentLeader {
		m.InformChannels(m.CurrentLeader)
	}
}

func (m *MonLeaderDetector) InformChannels(id int) {
	for _, channel := range m.LeaderChannels {
		channel <- id
	}
}
