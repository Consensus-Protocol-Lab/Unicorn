/*
 * Copyright (C) 2022 The Unicorn Authors
 * This file is part of The Unicorn library.
 *
 * The Unicorn is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The Unicorn is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The Unicorn.  If not, see <http://www.gnu.org/licenses/>.
 */

package interfaces

import (
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prysmaticlabs/prysm/v3/crypto/bls/common"
)

type BlsSigner interface {
	Sign(msg []byte) common.Signature
	AggregateSignatures(sigs []common.Signature) common.Signature
	FastAggregateVerify(pubKeys []common.PublicKey, hash common2.Hash) bool
	Marshal() []byte
	ConsenesusKeyFromBytes(priv []byte) (err error)
	StoreConsensusPublicKeyList(pubKeys []common.PublicKey)
	GetConsensusPublicKey(keyIndex int) ([]byte, error)
	StoreConsensusPublicKey(k []byte, pubKey []byte)
	VerifyValidatorSeal(header *types.Header, valSet ValidatorSet) error
}
