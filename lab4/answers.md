## Answers to Paxos Questions 

You should write down your answers to the
[questions](README.md#questions)
for Lab 4 in this file. 

1. Is it possible that Paxos enters an infinite loop? Explain.

Yes it is possible. It can happen if proposers tries to propose a new round with increasing round numbers at the same time. They will be upping the previous proposed round number and will be stuck there in an infinite loop. 

2. Is the value to agree on included in the Prepare message?

No, the value is not included in the prepare message. We want to agree on the current round number.

3. Does Paxos rely on an increasing proposal/round number in order to work? Explain.

Yes. When a node is preparing to become the leader, the round number is increased. When it becomes the leader, we are sure that this is the current leader and the previous leader will not be used as a leader from other nodes as the round number is not equal to the current round number. By increasing the round number we are sure that there are no more than one node that acts like a leader.

4. Look at this description for Phase 1B: If the proposal number N is larger than any previous proposal, then each Acceptor promises not to accept proposals less than N, and sends the value it last accepted for this instance to the Proposer. What is meant by "the value it last accepted"? And what is an "instance" in this case?

The value it last accepted is the last value the acceptor received in an accept message from the proposer. An instance is a node running the same paxos algorithm, meaning in the same network.

5. Explain, with an example, what will happen if there are multiple proposers.

6. What happens if two proposers both believe themselves to be the leader and send Prepare messages simultaneously?

If the prepare messages have the same round number it can either be one of them still reaching majority of promises and therefore will be the leader. Or none of the will get quorum. If they have different round numbers the highest of them will be chosen as leader.

7. What can we say about system synchrony if there are multiple proposers (or leaders)?

The algorithm is async, and does not rely on a global clock. 

8. Can an acceptor accept one value in round 1 and another value in round 2? Explain.

No. If the acceptor has promised to a proposer and received an accept from the proposer it cannot accept another value in round 2. But if we get the new message before the proposer sends an accept, it might be possible.

9. What must happen for a value to be "chosen"? What is the connection between chosen values and learned values?

For a value to be chosen:
    - Proposer gets a value from the clinet
    - Proposer send a prepare and get promise from the majority of the acceptors
    - Proposer sends the accept message to all acceptors and tell them to choose the value.
    - Learners will listen for chosen value
    - If the learner learns the value from the majority of nodes/acceptors
