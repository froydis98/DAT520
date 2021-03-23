package multipaxos

import (
	"dat520/lab5/bank"
	"fmt"
)

// Learner represents a learner as defined by the Multi-Paxos algorithm.
type Learner struct {
	ID           int
	NrOfNodes    int
	quorum       int
	decidedOut   chan<- DecidedValue
	learnIn      chan Learn
	stop         chan struct{}
	Val          Value
	Rnd          Round
	learnedSlots map[SlotID][]Learn
	lrnSent      map[SlotID]bool
}

// NewLearner returns a new Multi-Paxos learner. It takes the
// following arguments:
//
// id: The id of the node running this instance of a Paxos learner.
//
// nrOfNodes: The total number of Paxos nodes.
//
// decidedOut: A send only channel used to send values that has been learned,
// i.e. decided by the Paxos nodes.
func NewLearner(id int, nrOfNodes int, decidedOut chan<- DecidedValue) *Learner {
	return &Learner{
		ID:           id,
		NrOfNodes:    nrOfNodes,
		quorum:       (nrOfNodes / 2) + 1,
		decidedOut:   decidedOut,
		Val:          Value{ClientID: "0000", ClientSeq: -10, AccountNum: -1, Tnx: bank.Transaction{Op: -1, Amount: 0}},
		lrnSent:      make(map[SlotID]bool),
		Rnd:          Round(0),
		learnIn:      make(chan Learn),
		learnedSlots: make(map[SlotID][]Learn),
		stop:         make(chan struct{}),
	}
}

// Start starts l's main run loop as a separate goroutine. The main run loop
// handles incoming learn messages.
func (l *Learner) Start() {
	go func() {
		for {
			select {
			case LearnMessage := <-l.learnIn:
				val, slot, output := l.handleLearn(LearnMessage)
				if output {
					l.decidedOut <- DecidedValue{SlotID: slot, Value: val}
				}
			case <-l.stop:
				break
			}
		}
	}()
}

// Stop stops l's main run loop.
func (l *Learner) Stop() {
	l.stop <- struct{}{}
}

// DeliverLearn delivers learn lrn to learner l.
func (l *Learner) DeliverLearn(lrn Learn) {
	l.learnIn <- lrn
}

// Internal: handleLearn processes learn lrn according to the Multi-Paxos
// algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true, sid the id for the
// slot that was decided and val contain the decided value. If handleLearn
// returns false as output, then val and sid will have their zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, sid SlotID, output bool) {
	fmt.Println("INSIDE HANDLE LEARN, ROUND VALUES ARE ", learn.Rnd, l.Rnd)
	if learn.Rnd < l.Rnd {
		return val, sid, false
	} else if learn.Rnd == l.Rnd {
		if l.learnedSlots[learn.Slot] != nil {
			for _, learned := range l.learnedSlots[learn.Slot] {
				if learned.From == learn.From && learned.Rnd == learn.Rnd {
					fmt.Println("FAILED IF INSIDE THE FOR LOOP")
					return val, sid, false
				}
			}
			l.learnedSlots[learn.Slot] = append(l.learnedSlots[learn.Slot], learn)
		} else {
			l.learnedSlots[learn.Slot] = append(l.learnedSlots[learn.Slot], learn)
		}
	} else {
		l.Rnd = learn.Rnd
		l.learnedSlots[learn.Slot] = nil
		l.learnedSlots[learn.Slot] = append(l.learnedSlots[learn.Slot], learn)
	}
	if l.lrnSent[learn.Slot] {
		fmt.Println("learn has been sent already FAIL")
		return val, sid, false
	}
	fmt.Println("LEN IS: ", len(l.learnedSlots[learn.Slot]), "QOURUM IS: ", l.quorum)
	if len(l.learnedSlots[learn.Slot]) >= l.quorum {
		l.lrnSent[learn.Slot] = true
		fmt.Println("--------------------------- SUCCESS")

		return learn.Val, learn.Slot, true
	}
	fmt.Println("REACHED BOTTOM FAIL")

	return val, sid, false
}
