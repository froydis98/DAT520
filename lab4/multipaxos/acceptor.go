package multipaxos

import "sort"

// Acceptor represents an acceptor as defined by the Multi-Paxos algorithm.
type Acceptor struct { 
	id int
	rnd Round
	slotID SlotID
	slots map[SlotID]PromiseSlot
	promiseOut chan<- Promise
	prepareIn chan Prepare
	acceptIn chan Accept
	learnOut chan<- Learn
	stop chan bool
}

// NewAcceptor returns a new Multi-Paxos acceptor.
// It takes the following arguments:
//
// id: The id of the node running this instance of a Paxos acceptor.
//
// promiseOut: A send only channel used to send promises to other nodes.
//
// learnOut: A send only channel used to send learns to other nodes.
func NewAcceptor(id int, promiseOut chan<- Promise, learnOut chan<- Learn) *Acceptor {
	return &Acceptor{
		id: id,
		promiseOut: promiseOut,
		prepareIn: make(chan Prepare),
		acceptIn: make(chan Accept),
		learnOut: learnOut,
		stop: make(chan bool),
	}
}

// Start starts a's main run loop as a separate goroutine. The main run loop
// handles incoming prepare and accept messages.
func (a *Acceptor) Start() {
	go func() {
		for {
			select {
			case <-a.stop:
				return
			case prepare := <-a.prepareIn:
				promise, output := a.handlePrepare(prepare)
				if output {
					a.promiseOut <- promise
				}
			case accept := <-a.acceptIn:
				learn, output := a.handleAccept(accept)
				if output {
					a.learnOut <- learn
				}
			}
		}
	}()
}

// Stop stops a's main run loop.
func (a *Acceptor) Stop() {
	a.stop <- true
}

// DeliverPrepare delivers prepare prp to acceptor a.
func (a *Acceptor) DeliverPrepare(prp Prepare) {
	// TODO(student): distributed implementation
}

// DeliverAccept delivers accept acc to acceptor a.
func (a *Acceptor) DeliverAccept(acc Accept) {
	// TODO(student): distributed implementation
}

// Internal: handlePrepare processes prepare prp according to the Multi-Paxos
// algorithm. If handling the prepare results in acceptor a emitting a
// corresponding promise, then output will be true and prm contain the promise.
// If handlePrepare returns false as output, then prm will be a zero-valued
// struct.
func (a *Acceptor) handlePrepare(prp Prepare) (prm Promise, output bool) {
	if prp.Crnd > a.rnd {
		a.rnd = prp.Crnd
		a.slotID = prp.Slot
		accSlots := []PromiseSlot{}
		for id, slot := range a.slots {
			if id >= prp.Slot {
				accSlots = append(accSlots, slot)
			}
		}
		sort.SliceStable(accSlots, func(i, j int) bool {
			return accSlots[i].ID < accSlots[j].ID
		})
		promise := Promise{To: prp.From, From: a.id, Rnd: a.rnd}
		if len(accSlots) > 0 {
			promise.Slots = accSlots
		}
		return promise, true
	}
	return Promise{}, false
}

// Internal: handleAccept processes accept acc according to the Multi-Paxos
// algorithm. If handling the accept results in acceptor a emitting a
// corresponding learn, then output will be true and lrn contain the learn.  If
// handleAccept returns false as output, then lrn will be a zero-valued struct.
func (a *Acceptor) handleAccept(acc Accept) (lrn Learn, output bool) {
	// TODO(student): algorithm implementation
	return Learn{}, false
}

// TODO(student): Add any other unexported methods needed.
