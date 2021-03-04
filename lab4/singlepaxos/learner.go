package singlepaxos

// Learner represents a learner as defined by the single-decree Paxos
// algorithm.
type Learner struct { 
	id int
	NrOfNodes int
	Val Value
	Rnd Round
	Previous map[int]Value
}

// NewLearner returns a new single-decree Paxos learner. It takes the
// following arguments:
//
// id: The id of the node running this instance of a Paxos learner.
//
// nrOfNodes: The total number of Paxos nodes.
func NewLearner(id int, nrOfNodes int) *Learner {
	return &Learner{
		id: id,
		NrOfNodes: nrOfNodes,
		Val: ZeroValue,
		Rnd: NoRound,
		Previous: make(map[int]Value),
	}
}

// Internal: handleLearn processes learn lrn according to the single-decree
// Paxos algorithm. If handling the learn results in learner l emitting a
// corresponding decided value, then output will be true and val contain the
// decided value. If handleLearn returns false as output, then val will have
// its zero value.
func (l *Learner) handleLearn(learn Learn) (val Value, output bool) {
	if learn.Rnd >= l.Rnd {
		if (learn.Val != ZeroValue && learn.Val != l.Val) || l.Rnd != learn.Rnd {
			l.Previous = make(map[int]Value)
			l.Rnd = learn.Rnd
			l.Val = learn.Val
		}
		l.Previous[learn.From] = learn.Val
	}
	if len(l.Previous) > l.NrOfNodes/2 {
		l.Previous = make(map[int]Value)
		return l.Val, true
	}
	return ZeroValue, false
}
