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

package event

import (
	"github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
)

// RequestEvent is posted to propose a proposal (posting the incoming block to
// the main hotstuff engine anyway regardless of being the speaker or delegators)
type RequestEvent struct {
	Proposal interfaces.Proposal
}

// MessageEvent is posted for HotStuff engine communication (posting the incoming
// communication messages to the main hotstuff engine anyway)
type MessageEvent struct {
	Payload []byte
}

// FinalCommittedEvent is posted when a proposal is committed
type FinalCommittedEvent struct {
	Header *types.Header
}
