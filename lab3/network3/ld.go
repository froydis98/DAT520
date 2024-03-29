package main

import (
	"sort"
)

// A MonLeaderDetector represents a Monarchical Eventual Leader Detector as
// described at page 53 in:
// Christian Cachin, Rachid Guerraoui, and Luís Rodrigues: "Introduction to
// Reliable and Secure Distributed Programming" Springer, 2nd edition, 2011.
type MonLeaderDetector struct {
	currentLeader  int
	suspected      map[int]bool // map of node ids  considered suspected
	nodeIDs        []int
	leaderChannels []chan<- int
}

// NewMonLeaderDetector returns a new Monarchical Eventual Leader Detector
// given a list of node ids.
func NewMonLeaderDetector(nodeIDs []int) *MonLeaderDetector {
	m := &MonLeaderDetector{}
	suspected := make(map[int]bool)
	m.suspected = suspected
	sort.Ints(nodeIDs)
	if nodeIDs[len(nodeIDs)-1] < 0 {
		m.currentLeader = UnknownID
	} else {
		m.currentLeader = nodeIDs[len(nodeIDs)-1]
	}
	m.nodeIDs = nodeIDs
	return m
}

// Leader returns the current leader. Leader will return UnknownID if all nodes
// are suspected.
func (m *MonLeaderDetector) Leader() int {
	if m.currentLeader == UnknownID {
		return UnknownID
	}
	return m.currentLeader
}

// Suspect instructs the leader detector to consider the node with matching
// id as suspected. If the suspect indication result in a leader change
// the leader detector should publish this change to its subscribers.
func (m *MonLeaderDetector) Suspect(id int) {
	m.suspected[id] = true
	m.ChangeLeader()
}

// Restore instructs the leader detector to consider the node with matching
// id as restored. If the restore indication result in a leader change
// the leader detector should publish this change to its subscribers.
func (m *MonLeaderDetector) Restore(id int) {
	m.suspected[id] = false
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
	m.leaderChannels = append(m.leaderChannels, leaderChan)
	return leaderChan
}

func (m *MonLeaderDetector) ChangeLeader() {
	allTrue := false
	tempLeader := m.currentLeader
	for _, node := range m.nodeIDs {
		if !m.suspected[node] {
			if node < 0 {
				m.currentLeader = UnknownID
			} else {
				m.currentLeader = node
			}
			allTrue = true
		}
	}
	if !allTrue {
		m.currentLeader = UnknownID
	}
	if tempLeader != m.currentLeader {
		m.InformChannels(m.currentLeader)
	}
}

func (m *MonLeaderDetector) InformChannels(id int) {
	for _, channel := range m.leaderChannels {
		channel <- id
	}
}
