package core

import (
	"github.com/prysmaticlabs/prysm/v3/crypto/rand"
	blst "github.com/supranational/blst/bindings/go"
)

type BlsSigner struct {
	ConsensusKey *blst.SecretKey
	PublicKey    *blst.P1Affine
}

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

func NewBlsSigner(consensusKey *blst.SecretKey) *BlsSigner {
	return &BlsSigner{
		ConsensusKey: consensusKey,
		PublicKey:    new(blst.P1Affine).From(consensusKey),
	}
}

func (b *BlsSigner) Sign(msg []byte) *blst.P2Affine {
	return new(blst.P2Affine).Sign(b.ConsensusKey, msg, dst)
}

func (b *BlsSigner) AggregateSignatures(sigs []*blst.P2Affine) *blst.P2Affine {
	if len(sigs) == 0 {
		return nil
	}

	rawSigs := make([]*blst.P2Affine, len(sigs))
	for i := 0; i < len(sigs); i++ {
		rawSigs[i] = sigs[i]
	}

	// Signature and PKs are assumed to have been validated upon decompression!
	signature := new(blst.P2Aggregate)
	signature.Aggregate(rawSigs, false)
	return signature.ToAffine()
}

func (b *BlsSigner) FastAggregateVerify(aggregatedSig *blst.P2Affine, publicKeys []*blst.P1Affine, msg []byte) bool {
	if len(publicKeys) == 0 {
		return false
	}
	rawKeys := make([]*blst.P1Affine, len(publicKeys))
	for i := 0; i < len(publicKeys); i++ {
		rawKeys[i] = publicKeys[i]
	}
	return aggregatedSig.FastAggregateVerify(true, rawKeys, msg[:], dst)
}

func generateKey() *blst.SecretKey {
	// Generate 32 bytes of randomness
	var ikm [32]byte
	_, err := rand.NewGenerator().Read(ikm[:])
	if err != nil {
		return nil
	}
	// Defensive check, that we have not generated a secret key,
	secKey := blst.KeyGen(ikm[:])
	return secKey
}
