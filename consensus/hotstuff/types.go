package hotstuff

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"math/big"
)

// Proposal supports retrieving height and serialized block to be used during HotStuff consensus.
// It is the interface that abstracts different message structure. (consensus/hotstuff/core/core.go)
type Proposal interface {
	// Number retrieves the block height number of this proposal.
	Number() *big.Int

	// Hash retrieves the hash of this proposal.
	Hash() common.Hash

	ParentHash() common.Hash

	Coinbase() common.Address

	Time() uint64

	EncodeRLP(w io.Writer) error

	DecodeRLP(s *rlp.Stream) error
}

type Request struct {
	Proposal Proposal
}

type MsgType interface {
	String() string
	Value() uint64
}

type View struct {
	Round  *big.Int
	Height *big.Int
}

type Message struct {
	Code          MsgType
	View          *View
	Msg           []byte
	Address       common.Address
	Signature     []byte
	CommittedSeal []byte
}
