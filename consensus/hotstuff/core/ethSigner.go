package core

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"strconv"
)

type EthSigner struct {
	address     common.Address
	privateKey  *ecdsa.PrivateKey
	db          ethdb.Database
	ValidatorNo int
}

var ValidatorAddressPrefix = "bls-public-key"

func NewEthSigner(privateKey *ecdsa.PrivateKey, db ethdb.Database) *EthSigner {
	return &EthSigner{
		address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		privateKey: privateKey,
		db:         db,
	}
}

func (ethSigner *EthSigner) Address() common.Address {
	return ethSigner.address
}

// SigHash returns the hash which is used as input for the Hotstuff
// signing. It is the hash of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func (ethSigner *EthSigner) SigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	// Clean seal is required for calculating proposer seal.
	rlp.Encode(hasher, types.HotstuffFilteredHeader(header, false))
	hasher.Sum(hash[:0])
	return hash
}

func (ethSigner *EthSigner) Sign(data []byte) ([]byte, error) {
	hashData := crypto.Keccak256(data)
	return crypto.Sign(hashData, ethSigner.privateKey)
}

// Recover extracts the proposer address from a signed header.
func (ethSigner *EthSigner) Recover(header *types.Header) (common.Address, error) {

	// Retrieve the signature from the header extra-data
	extra, err := types.ExtractHotstuffExtra(header)
	if err != nil {
		return common.Address{}, errInvalidExtraDataFormat
	}

	payload := ethSigner.SigHash(header).Bytes()
	addr, err := getSignatureAddress(payload, extra.LeaderSeal)
	if err != nil {
		return addr, err
	}

	return addr, nil
}

func (ethSigner *EthSigner) VerifyLeaderSeal(header *types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}

	// resolve the authorization key and check against signers
	signer, err := ethSigner.Recover(header)
	if err != nil {
		return err
	}
	if signer != header.Coinbase {
		return errInvalidSigner
	}
	return nil
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

func (ethSigner *EthSigner) StoreValidator(val common.Address, valIndex int) {
	ethSigner.db.Put(append([]byte(ValidatorAddressPrefix), []byte(strconv.Itoa(valIndex))...), val.Bytes())
}

func (ethSigner *EthSigner) GetValidator(valIndex int) (common.Address, error) {
	val, err := ethSigner.db.Get(append([]byte(ValidatorAddressPrefix), []byte(strconv.Itoa(valIndex))...))
	return common.BytesToAddress(val), err
}

// GetSignatureAddress gets the address address from the signature
func getSignatureAddress(data []byte, sig []byte) (common.Address, error) {
	// 1. Keccak data
	hashData := crypto.Keccak256(data)
	// 2. Recover public key
	pubkey, err := crypto.SigToPub(hashData, sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}
