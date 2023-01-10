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

package validator

import (
	"errors"
	. "github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

var ErrInvalidParticipant = errors.New("invalid participants")

type defaultValidator struct {
	address common.Address
}

func (val *defaultValidator) Address() common.Address {
	return val.address
}

func (val *defaultValidator) String() string {
	return val.Address().String()
}

// ----------------------------------------------------------------------------

type defaultSet struct {
	validators Validators
	policy     SelectProposerPolicy

	proposer    Validator
	validatorMu sync.RWMutex
	selector    ProposalSelector
}

func newDefaultSet(addrs []common.Address, policy SelectProposerPolicy) *defaultSet {
	valSet := &defaultSet{}

	valSet.policy = policy
	// init validators
	valSet.validators = make([]Validator, len(addrs))
	for i, addr := range addrs {
		valSet.validators[i] = New(addr)
	}
	// sort validator
	sort.Sort(valSet.validators)
	// init proposer
	if valSet.Size() > 0 {
		valSet.proposer = valSet.GetByIndex(0)
	}
	valSet.selector = roundRobinSelector
	if policy == Sticky {
		valSet.selector = stickySelector
	}
	if policy == VRF {
		valSet.selector = vrfSelector
	}

	return valSet
}

func (valSet *defaultSet) CalcProposer(lastProposer common.Address, round uint64) {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	valSet.proposer = valSet.selector(valSet, lastProposer, round)
}

func (valSet *defaultSet) CalcProposerByIndex(index uint64) {
	if index > 1 {
		index = (index - 1) % uint64(len(valSet.validators))
	} else {
		index = 0
	}
	valSet.proposer = valSet.validators[index]
}

func (valSet *defaultSet) Size() int {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	return len(valSet.validators)
}

func (valSet *defaultSet) List() []Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	return valSet.validators
}

func (valSet *defaultSet) AddressList() []common.Address {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()

	vals := make([]common.Address, valSet.Size())
	for i, v := range valSet.List() {
		vals[i] = v.Address()
	}
	return vals
}

func (valSet *defaultSet) GetByIndex(i uint64) Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	if i < uint64(valSet.Size()) {
		return valSet.validators[i]
	}
	return nil
}

func (valSet *defaultSet) GetByAddress(addr common.Address) (int, Validator) {
	for i, val := range valSet.List() {
		if addr == val.Address() {
			return i, val
		}
	}
	return -1, nil
}

func (valSet *defaultSet) GetProposer() Validator {
	return valSet.proposer
}

func (valSet *defaultSet) IsProposer(address common.Address) bool {
	_, val := valSet.GetByAddress(address)
	return reflect.DeepEqual(valSet.GetProposer(), val)
}

func (valSet *defaultSet) AddValidator(address common.Address) bool {
	valSet.validatorMu.Lock()
	defer valSet.validatorMu.Unlock()
	for _, v := range valSet.validators {
		if v.Address() == address {
			return false
		}
	}
	valSet.validators = append(valSet.validators, New(address))
	// TODO: we may not need to re-sort it again
	// sort validator
	sort.Sort(valSet.validators)
	return true
}

func (valSet *defaultSet) RemoveValidator(address common.Address) bool {
	valSet.validatorMu.Lock()
	defer valSet.validatorMu.Unlock()

	for i, v := range valSet.validators {
		if v.Address() == address {
			valSet.validators = append(valSet.validators[:i], valSet.validators[i+1:]...)
			return true
		}
	}
	return false
}

func (valSet *defaultSet) Copy() ValidatorSet {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()

	addresses := make([]common.Address, 0, len(valSet.validators))
	for _, v := range valSet.validators {
		addresses = append(addresses, v.Address())
	}
	return NewSet(addresses, valSet.policy)
}

func (valSet *defaultSet) ParticipantsNumber(list []common.Address) int {
	if list == nil || len(list) == 0 {
		return 0
	}
	size := 0
	for _, v := range list {
		if index, _ := valSet.GetByAddress(v); index < 0 {
			continue
		} else {
			size += 1
		}
	}
	return size
}

func (valSet *defaultSet) CheckQuorum(committers []common.Address) error {
	validators := valSet.Copy()
	validSeal := 0
	for _, addr := range committers {
		if validators.RemoveValidator(addr) {
			validSeal++
			continue
		}
		return ErrInvalidParticipant
	}

	// The length of validSeal should be larger than number of faulty node + 1
	if validSeal <= validators.Q() {
		return ErrInvalidParticipant
	}
	return nil
}

func (valSet *defaultSet) F() int { return int(math.Ceil(float64(valSet.Size())/3)) - 1 }

func (valSet *defaultSet) Q() int { return int(math.Ceil(float64(2*valSet.Size()) / 3)) }

func (valSet *defaultSet) Policy() SelectProposerPolicy { return valSet.policy }

func (valSet *defaultSet) Cmp(src ValidatorSet) bool {
	n := valSet.ParticipantsNumber(src.AddressList())
	if n != valSet.Size() || n != src.Size() {
		return false
	}
	return true
}

func calcSeed(valSet ValidatorSet, proposer common.Address, round uint64) uint64 {
	offset := 0
	if idx, val := valSet.GetByAddress(proposer); val != nil {
		offset = idx
	}
	return uint64(offset) + round
}

func emptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

func roundRobinSelector(valSet ValidatorSet, proposer common.Address, round uint64) Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if emptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(valSet, proposer, round) + 1
	}
	pick := seed % uint64(valSet.Size())
	return valSet.GetByIndex(pick)
}

func stickySelector(valSet ValidatorSet, proposer common.Address, round uint64) Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if emptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(valSet, proposer, round)
	}
	pick := seed % uint64(valSet.Size())
	return valSet.GetByIndex(pick)
}

// TODO: implement VRF
func vrfSelector(valSet ValidatorSet, proposer common.Address, round uint64) Validator {
	return nil
}
