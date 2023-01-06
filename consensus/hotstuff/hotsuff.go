package hotstuff

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/consensus"
	hotstuffEngine "github.com/ethereum/go-ethereum/consensus/hotstuff/engine"
	blst "github.com/supranational/blst/bindings/go"
)

func New(privateKey *ecdsa.PrivateKey, consensusKey *blst.SecretKey) consensus.Hotstuff {
	return hotstuffEngine.New(privateKey, consensusKey)
}
