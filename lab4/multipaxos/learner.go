package multipaxos

// Learner represents a learner as defined by the Multi-Paxos algorithm.
type Learner struct { // TODO(student): algorithm and distributed implementation
	ID         int
	NrOfNodes  int
	decidedOut chan<- DecidedValue
	learnIn    chan Learn
	stop       chan struct{}
	Val        Value
	Rnd        Round
	Previous   map[int]Value
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
		ID:         id,
		NrOfNodes:  nrOfNodes,
		decidedOut: decidedOut,
		Val:        Value{ClientID: "0000", ClientSeq: -10, Command: "none"},
		Rnd:        Round(0),
	}
}

// Start starts l's main run loop as a separate goroutine. The main run loop
// handles incoming learn messages.
func (l *Learner) Start() {
	go func() {
		for {
			// TODO(student): distributed implementation
			select {
			case LearnMessage := <-l.learnIn:
				l.handleLearn(LearnMessage)
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
	// TODO(student): distributed implementation
}

// Internal: handleLearn processes learn lrn according to the Multi-Paxos
// algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true, sid the id for the
// slot that was decided and val contain the decided value. If handleLearn
// returns false as output, then val and sid will have their zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, sid SlotID, output bool) {
	if learn.Rnd >= l.Rnd {
		if (learn.Val != l.Val) || l.Rnd != learn.Rnd {
			l.Previous = make(map[int]Value)
			l.Rnd = learn.Rnd
			l.Val = learn.Val
		}
		l.Previous[learn.From] = learn.Val
	}
	if len(l.Previous) > l.NrOfNodes/2 {
		l.Previous = make(map[int]Value)
		return l.Val, 1, true
	}
	return Value{ClientID: "-1", ClientSeq: -1, Command: "-1"}, -1, false
}
