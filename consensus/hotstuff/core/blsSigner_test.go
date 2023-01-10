package core

import (
	"github.com/ethereum/go-ethereum/crypto"
	blst "github.com/prysmaticlabs/prysm/v3/crypto/bls/blst"
	"github.com/prysmaticlabs/prysm/v3/crypto/bls/common"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestBlsSignVerify(t *testing.T) {
	sk, _ := generateKey()
	blsSigner := NewBlsSigner(&sk, nil)
	msg := []byte{0x11, 0x22}
	msgHash := crypto.Keccak256(msg)
	t.Log(len(msgHash))
	//wrongMsg := []byte{0x22, 0x22}
	sig := blsSigner.Sign(msg)
	t.Log(len(sig.Marshal()))
	t.Log(reflect.TypeOf(sig))
	//t.Log(sig.Verify(false, BlsSigner.ConsensusPublicKey, false, msg, dst))
	//t.Log(sig.Verify(false, BlsSigner.ConsensusPublicKey, false, wrongMsg, dst))
}

func TestBlsAggregateSignVerify(t *testing.T) {

	msg := [32]byte{0x11, 0x22}
	wrongMsg := [32]byte{0x22, 0x22}
	keyGen := func() *common.SecretKey {
		k, _ := blst.RandKey()
		return &k
	}
	testSuite := []struct {
		signer *BlsSigner
	}{
		{
			NewBlsSigner(keyGen(), nil),
		},
		{
			NewBlsSigner(keyGen(), nil),
		},
	}
	sigs := make([]common.Signature, len(testSuite))
	pubKeys := make([]common.PublicKey, len(testSuite))
	for i, testCase := range testSuite {
		sig := testCase.signer.Sign(msg[:])
		sigs[i] = sig
		pubKeys[i] = *testCase.signer.ConsensusPublicKey
	}
	testSuite[0].signer.aggSignatures = testSuite[0].signer.AggregateSignatures(sigs)
	res := testSuite[0].signer.FastAggregateVerify(pubKeys, msg)
	assert.Equal(t, true, res, "did not verify")
	wrongRes := testSuite[0].signer.FastAggregateVerify(pubKeys, wrongMsg)
	assert.Equal(t, false, wrongRes, "verify passed")
}
