package core

import (
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	blst "github.com/prysmaticlabs/prysm/v3/crypto/bls/blst"
	"github.com/prysmaticlabs/prysm/v3/crypto/bls/common"
	"strconv"
)

type BlsSigner struct {
	ConsensusKey       *common.SecretKey
	ConsensusPublicKey *common.PublicKey
	db                 ethdb.Database
	ValidatorNo        int
	aggSignatures      common.Signature
}

var (
	dst                      = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")
	ConsensusPublicKeyPrefix = "bls-public-key"
)

func NewBlsSigner(consensusKey *common.SecretKey, db ethdb.Database) *BlsSigner {
	pk := (*consensusKey).PublicKey()
	return &BlsSigner{
		ConsensusKey:       consensusKey,
		ConsensusPublicKey: &pk,
		db:                 db,
	}

}

func generateKey() (common.SecretKey, error) {
	return blst.RandKey()
}

func (blsSigner *BlsSigner) Sign(msg []byte) common.Signature {
	return (*blsSigner.ConsensusKey).Sign(msg)
}

func (blsSigner *BlsSigner) AggregateSignatures(sigs []common.Signature) common.Signature {
	return blst.AggregateSignatures(sigs)
}

func (blsSigner *BlsSigner) FastAggregateVerify(pubKeys []common.PublicKey, msg common2.Hash) bool {
	return blsSigner.aggSignatures.FastAggregateVerify(pubKeys, msg)
}

// Marshal a secret key into a LittleEndian byte slice.
func (blsSigner *BlsSigner) Marshal() []byte {
	return (*blsSigner.ConsensusKey).Marshal()
}

// Marshal a secret key into a LittleEndian byte slice.
func (blsSigner *BlsSigner) ConsenesusKeyFromBytes(priv []byte) (err error) {
	*blsSigner.ConsensusKey, err = blst.SecretKeyFromBytes(priv)
	return err
}

func (blsSigner *BlsSigner) StoreConsensusPublicKeyList(pubKeys []common.PublicKey) {
	for k, pubKey := range pubKeys {
		blsSigner.StoreConsensusPublicKey([]byte(strconv.Itoa(k)), pubKey.Marshal())
	}
}

func (blsSigner *BlsSigner) GetConsensusPublicKey(keyIndex int) ([]byte, error) {
	storeKey := append([]byte(ConsensusPublicKeyPrefix), []byte(strconv.Itoa(keyIndex))...)
	return blsSigner.db.Get(storeKey)
}

func (blsSigner *BlsSigner) StoreConsensusPublicKey(k []byte, pubKey []byte) {
	blsSigner.db.Put(append([]byte(ConsensusPublicKeyPrefix), k...), pubKey)
}

func (blsSigner *BlsSigner) VerifyValidatorSeal(header *types.Header, valSet interfaces.ValidatorSet) error {

	extra, err := types.ExtractHotstuffExtra(header)
	if err != nil {
		return errInvalidExtraDataFormat
	}

	// The length of Committed seals should be larger than 0
	if len(extra.AggregatedValidatorsSeal) == 0 {
		return errEmptyCommittedSeals
	}

	ValidatorList := extra.ParticipantsIndex
	if len(ValidatorList) < valSet.Q() {
		return errInvalidValidatorSeals
	}

	pubkeys := make([]common.PublicKey, len(ValidatorList))
	for i := 0; i < len(ValidatorList); i++ {
		pk, err := blsSigner.GetConsensusPublicKey(i)
		pubKey, err := blst.PublicKeyFromBytes(pk)
		if err != nil {
			return errValidatorNotFound
		}
		pubkeys[i] = pubKey
	}

	if blsSigner.aggSignatures, err = blst.SignatureFromBytes(extra.AggregatedValidatorsSeal); err != nil {
		return err
	}

	if res := blsSigner.FastAggregateVerify(pubkeys, header.Hash()); res {
		return nil
	}
	return errInvalidValidatorSeals
}
