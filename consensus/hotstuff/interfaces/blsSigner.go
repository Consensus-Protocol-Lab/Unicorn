package interfaces

import (
	blst "github.com/supranational/blst/bindings/go"
)

type BlsPacker interface {
	Sign(msg []byte) *blst.P2Affine
	AggregateSignatures(sigs []*blst.P2Affine) *blst.P2Affine
	// this is to verify signatures of the same msgs
	FastAggregateVerify(aggregatedSig *blst.P2Affine, publicKeys []*blst.P1Affine, msg []byte) bool
}
