package singlepaxos

// Proposer represents a proposer as defined by the single-decree Paxos
// algorithm.
type Proposer struct {
	id          int
	crnd        Round
	clientValue Value
	A           map[*Acceptor]bool
	nrOfNodes   int
	MV          map[int]Promise
}

// NewProposer returns a new single-decree Paxos proposer.
// It takes the following arguments:
//
// id: The id of the node running this instance of a Paxos proposer.
//
// nrOfNodes: The total number of Paxos nodes.
//
// The proposer's internal crnd field should initially be set to the value of
// its id.
func NewProposer(id int, nrOfNodes int) *Proposer {
	return &Proposer{
		id:          id,
		crnd:        Round(id),
		A:           make(map[*Acceptor]bool),
		clientValue: ZeroValue,
		nrOfNodes:   nrOfNodes,
		MV:          make(map[int]Promise),
	}
}

// Internal: handlePromise processes promise prm according to the single-decree
// Paxos algorithm. If handling the promise results in proposer p emitting a
// corresponding accept, then output will be true and acc contain the promise.
// If handlePromise returns false as output, then acc will be a zero-valued
// struct.
func (p *Proposer) handlePromise(prm Promise) (acc Accept, output bool) {
	p.MV[prm.From] = prm
	p.crnd = prm.Rnd
	if len(p.MV) > (p.nrOfNodes / 2) {
		var temp Promise
		for _, promise := range p.MV {
			if promise.Vrnd > temp.Vrnd {
				temp = promise
			}
		}
		if temp.Vval != ZeroValue {
			p.clientValue = temp.Vval
		}
		if p.clientValue != ZeroValue {
			if (p.crnd == NoRound) || (p.crnd == 0) {
				return Accept{}, false
			}
			return Accept{From: p.id, Rnd: p.crnd, Val: p.clientValue}, true
		}
	}
	return Accept{}, false
}

// Internal: increaseCrnd increases proposer p's crnd field by the total number
// of Paxos nodes.
func (p *Proposer) increaseCrnd() {
	p.crnd += Round(p.nrOfNodes)
}
