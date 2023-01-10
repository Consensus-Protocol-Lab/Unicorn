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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthSigner interface {
	Address() common.Address
	SigHash(header *types.Header) (hash common.Hash)
	Sign(data []byte) ([]byte, error)
	SealBeforeCommit(h *types.Header) error
	StoreValidator(val common.Address, valIndex int)
	GetValidator(valIndex int) (common.Address, error)
}
