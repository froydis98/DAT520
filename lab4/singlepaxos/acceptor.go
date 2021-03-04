package singlepaxos

// Acceptor represents an acceptor as defined by the single-decree Paxos
// algorithm.
type Acceptor struct {
	id int
	rnd Round
	Vrnd Round
	Vval Value
	P map[*Proposer]bool
	L map[*Learner]bool
}

// NewAcceptor returns a new single-decree Paxos acceptor.
// It takes the following arguments:
//
// id: The id of the node running this instance of a Paxos acceptor.
func NewAcceptor(id int) *Acceptor {
	return &Acceptor{
		id: id,
		rnd: 0,
		Vrnd: NoRound,
		Vval: ZeroValue,
		P: make(map[*Proposer]bool),
		L: make(map[*Learner]bool),
	}
}

// Internal: handlePrepare processes prepare prp according to the single-decree
// Paxos algorithm. If handling the prepare results in acceptor a emitting a
// corresponding promise, then output will be true and prm contain the promise.
// If handlePrepare returns false as output, then prm will be a zero-valued
// struct.
func (a *Acceptor) handlePrepare(prp Prepare) (prm Promise, output bool) {
	if prp.Crnd > a.rnd {
		a.rnd = prp.Crnd
		return Promise{To: prp.From, From: a.id, Rnd: a.rnd, Vrnd: a.Vrnd, Vval: a.Vval}, true
	}
	return Promise{}, false
}

// Internal: handleAccept processes accept acc according to the single-decree
// Paxos algorithm. If handling the accept results in acceptor a emitting a
// corresponding learn, then output will be true and lrn contain the learn.  If
// handleAccept returns false as output, then lrn will be a zero-valued struct.
func (a *Acceptor) handleAccept(acc Accept) (lrn Learn, output bool) {
	if acc.Rnd >= a.rnd && acc.Rnd != a.Vrnd {
		a.rnd = acc.Rnd
		a.Vrnd = acc.Rnd
		a.Vval = acc.Val
		return Learn{From: a.id, Rnd: a.rnd, Val: a.Vval}, true
	}
	return Learn{}, false
}
