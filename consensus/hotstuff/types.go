package hotstuff

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"
	"math/big"
)

type Request struct {
	Proposal interfaces.Proposal
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
