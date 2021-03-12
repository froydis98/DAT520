package multipaxos

import "fmt"

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
		Val:          Value{ClientID: "0000", ClientSeq: -10, Command: "none"},
		Rnd:          Round(0),
		learnIn:      make(chan Learn),
		learnedSlots: make(map[SlotID][]Learn),
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
				return
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
		fmt.Println("FAILED FIRST TEST")
		return val, sid, false
	} else if learn.Rnd == l.Rnd {
		fmt.Println("INSIDE THE ELSE IF")

		if l.learnedSlots[learn.Slot] == nil {
			fmt.Println("INSIDE THE FIRST IF INSIDE ELSE IF")

			l.learnedSlots[learn.Slot] = []Learn{}
		}
		for _, learned := range l.learnedSlots[learn.Slot] {
			fmt.Println("WE ARE IN THE FOR LOOP")

			if learned.From == learn.From && learned.Rnd == learn.Rnd && learned.Slot == learn.Slot {
				fmt.Println("FAILED IF INSIDE THE FOR LOOP")

				return val, sid, false
			}
		}
		l.learnedSlots[learn.Slot] = append(l.learnedSlots[learn.Slot], learn)
		if len(l.learnedSlots[learn.Slot]) == l.quorum {
			fmt.Println("all learned slots: ", l.learnedSlots)
			return learn.Val, learn.Slot, true
		}
	} else {
		fmt.Println("INSIDE THE ELSE ")
		l.Rnd = learn.Rnd
		l.learnedSlots = map[SlotID][]Learn{}
		l.learnedSlots[learn.Slot] = []Learn{learn}
	}
	return val, sid, false
}
