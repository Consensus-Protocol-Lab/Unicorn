package core

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

type EthSigner struct {
	address    common.Address
	privateKey *ecdsa.PrivateKey
}

func NewEthSigner(privateKey *ecdsa.PrivateKey) *EthSigner {
	return &EthSigner{
		address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		privateKey: privateKey,
	}
}

// SigHash returns the hash which is used as input for the Hotstuff
// signing. It is the hash of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func (e *EthSigner) SigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	// Clean seal is required for calculating proposer seal.
	rlp.Encode(hasher, types.HotstuffFilteredHeader(header, false))
	hasher.Sum(hash[:0])
	return hash
}

func (e *EthSigner) Sign(data []byte) ([]byte, error) {
	hashData := crypto.Keccak256(data)
	return crypto.Sign(hashData, e.privateKey)
}

// SignerSeal proposer sign the header hash and fill extra seal with signature.
func (ethSigner *EthSigner) SealBeforeCommit(h *types.Header) error {
	sigHash := ethSigner.SigHash(h)
	seal, err := ethSigner.Sign(sigHash.Bytes())
	if err != nil {
		return errInvalidSignature
	}

	if len(seal)%types.HotstuffExtraSeal != 0 {
		return errInvalidSignature
	}

	extra, err := types.ExtractHotstuffExtra(h)
	if err != nil {
		return err
	}
	extra.LeaderSeal = seal
	payload, err := rlp.EncodeToBytes(&extra)
	if err != nil {
		return err
	}
	h.Extra = append(h.Extra[:types.HotstuffExtraVanity], payload...)
	return nil
}
