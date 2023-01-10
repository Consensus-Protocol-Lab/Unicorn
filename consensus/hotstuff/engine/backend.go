package engine

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// Address returns the owner's address
func (e *HotStuffEngine) Address() common.Address {
	return e.signer.EthSigner.Address()
}

// EventMux returns the event mux in backend
func (e *HotStuffEngine) EventMux() *event.TypeMux {
	return e.eventMux
}

//
//// Broadcast sends a message to all validators (include self)
//func (e *HotStuffEngine) Broadcast(valSet ValidatorSet, payload []byte) error {
//	// send to others
//	if err := e.Gossip(valSet, payload); err != nil {
//		return err
//	}
//	// send to self
//	msg := event2.MessageEvent{
//		Payload: payload,
//	}
//	go e.EventMux().Post(msg)
//	return nil
//}
//
//// Gossip sends a message to all validators (exclude self)
//func (e *HotStuffEngine) Gossip(valSet ValidatorSet, payload []byte) error {
//	hash := hotstuff.RLPHash(payload)
//	e.knownMessages.Add(hash, true)
//
//	targets := make(map[common.Address]bool)
//	for _, val := range valSet.List() { // hotstuff/validator/default.go - defaultValidator
//		if val.Address() != e.Address() {
//			targets[val.Address()] = true
//		}
//	}
//	if e.broadcaster != nil && len(targets) > 0 {
//		ps := e.broadcaster.FindPeers(targets)
//		for addr, p := range ps {
//			ms, ok := e.recentMessages.Get(addr)
//			var m *lru.ARCCache
//			if ok {
//				m, _ = ms.(*lru.ARCCache)
//				if _, k := m.Get(hash); k {
//					// This peer had this event, skip it
//					continue
//				}
//			} else {
//				m, _ = lru.NewARC(inmemoryMessages)
//			}
//
//			m.Add(hash, true)
//			e.recentMessages.Add(addr, m)
//			go p.Send(hotstuffMsg, payload)
//		}
//	}
//	return nil
//
//}
//
//// Unicast send a message to single peer
//func (e *HotStuffEngine) Unicast(valSet ValidatorSet, payload []byte) error {
//
//}
//
//// PreCommit write seal to header and assemble new qc
//func (e *HotStuffEngine) PreCommit(proposal Proposal, seals [][]byte) (Proposal, error) {
//
//}
//
//// ForwardCommit assemble unsealed block and sealed extra into an new full block
//func (e *HotStuffEngine) ForwardCommit(proposal Proposal, extra []byte) (Proposal, error) {
//
//}
//
//// Commit delivers an approved proposal to backend.
//// The delivered proposal will be put into blockchain.
//func (e *HotStuffEngine) Commit(proposal Proposal) error {
//
//}
//
//// Verify verifies the proposal. If a consensus.ErrFutureBlock error is returned,
//// the time difference of the proposal and current time is also returned.
//func (e *HotStuffEngine) Verify(Proposal) (time.Duration, error) {
//
//}
//
//// Verify verifies the proposal. If a consensus.ErrFutureBlock error is returned,
//// the time difference of the proposal and current time is also returned.
//func (e *HotStuffEngine) VerifyUnsealedProposal(Proposal) (time.Duration, error) {
//
//}
//
//// LastProposal retrieves latest committed proposal and the address of proposer
//func (e *HotStuffEngine) LastProposal() (Proposal, common.Address) {
//
//}
//
//// HasBadBlock returns whether the block with the hash is a bad block
//func (e *HotStuffEngine) HasBadProposal(hash common.Hash) bool {
//
//}
//
//// ValidateBlock execute block which contained in prepare message, and validate block state
//func (e *HotStuffEngine) ValidateBlock(block *types.Block) error {
//
//}
