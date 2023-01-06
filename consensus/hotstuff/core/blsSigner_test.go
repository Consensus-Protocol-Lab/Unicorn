package core

import (
	"github.com/stretchr/testify/assert"
	blst "github.com/supranational/blst/bindings/go"
	"testing"
)

func TestBlsSignVerify(t *testing.T) {

	blsSigner := NewBlsSigner(generateKey())
	msg := []byte{0x11, 0x22}
	wrongMsg := []byte{0x22, 0x22}
	sig := blsSigner.Sign(msg)
	t.Log(sig.Verify(false, blsSigner.PublicKey, false, msg, dst))
	t.Log(sig.Verify(false, blsSigner.PublicKey, false, wrongMsg, dst))
}

func TestBlsAggregateSignVerify(t *testing.T) {

	msg := []byte{0x11, 0x22}
	wrongMsg := []byte{0x22, 0x22}
	testSuite := []struct {
		signer *BlsSigner
	}{
		{
			NewBlsSigner(generateKey()),
		},
		{
			NewBlsSigner(generateKey()),
		},
	}
	sigs := make([]*blst.P2Affine, len(testSuite))
	pubKeys := make([]*blst.P1Affine, len(testSuite))
	for i, testCase := range testSuite {
		sig := testCase.signer.Sign(msg)
		sigs[i] = sig
		pubKeys[i] = testCase.signer.PublicKey
	}
	aggSig := testSuite[0].signer.AggregateSignatures(sigs)
	res := testSuite[0].signer.FastAggregateVerify(aggSig, pubKeys, msg)
	assert.Equal(t, true, res, "verify passed")
	wrongRes := testSuite[0].signer.FastAggregateVerify(aggSig, pubKeys, wrongMsg)
	assert.Equal(t, false, wrongRes, "verify failed")
}
