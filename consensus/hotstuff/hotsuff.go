package hotstuff

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/config"
	hotstuffEngine "github.com/ethereum/go-ethereum/consensus/hotstuff/engine"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/prysmaticlabs/prysm/v3/crypto/bls/common"
)

func New(privateKey *ecdsa.PrivateKey, consensusKey *common.SecretKey, config *config.Config, db ethdb.Database) consensus.Hotstuff {
	return hotstuffEngine.New(privateKey, consensusKey, config, db)
}
