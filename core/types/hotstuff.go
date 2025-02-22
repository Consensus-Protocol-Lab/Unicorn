// Copyright 2022 The Unicorn Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
)

var (
	// HotstuffDigest represents a hash of "Hotstuff practical byzantine fault tolerance"
	// to identify whether the block is from Istanbul consensus engine
	HotstuffDigest = common.HexToHash("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365")

	HotstuffExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
	HotstuffExtraSeal   = 65 // Fixed number of extra-data bytes reserved for validator seal

	// ErrInvalidHotstuffHeaderExtra is returned if the length of extra-data is less than 32 bytes
	ErrInvalidHotstuffHeaderExtra = errors.New("invalid istanbul header extra-data")
)

type HotstuffExtra struct {
	Validators               []common.Address // consensus participants address of the current block
	LeaderSeal               []byte           // proposer seal
	AggregatedValidatorsSeal []byte           // aggregated 2f+1 validators' seals
	ParticipantsIndex        []int            // index of 2f+1 validators that participate in the seal process
	Salt                     []byte           // omit empty
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (ist *HotstuffExtra) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		ist.Validators,
		ist.LeaderSeal,
		ist.AggregatedValidatorsSeal,
		ist.Salt,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (ist *HotstuffExtra) DecodeRLP(s *rlp.Stream) error {
	extra := new(HotstuffExtra)
	if err := s.Decode(&extra); err != nil {
		return err
	}
	ist.Validators, ist.LeaderSeal, ist.AggregatedValidatorsSeal, ist.Salt = extra.Validators, extra.LeaderSeal, extra.AggregatedValidatorsSeal, extra.Salt
	return nil
}

// ExtractHotstuffExtra extracts all values of the HotstuffExtra from the header. It returns an
// error if the length of the given extra-data is less than 32 bytes or the extra-data can not
// be decoded.
func ExtractHotstuffExtra(h *Header) (*HotstuffExtra, error) {
	return ExtractHotstuffExtraPayload(h.Extra)
}

func ExtractHotstuffExtraPayload(extra []byte) (*HotstuffExtra, error) {
	if len(extra) < HotstuffExtraVanity {
		return nil, ErrInvalidHotstuffHeaderExtra
	}

	var hotstuffExtra *HotstuffExtra
	err := rlp.DecodeBytes(extra[HotstuffExtraVanity:], &hotstuffExtra)
	if err != nil {
		return nil, err
	}
	return hotstuffExtra, nil
}

// HotstuffFilteredHeader returns a filtered header which some information (like seal, committed seals)
// are clean to fulfill the Istanbul hash rules. It returns nil if the extra-data cannot be
// decoded/encoded by rlp.
func HotstuffFilteredHeader(h *Header, keepSeal bool) *Header {
	newHeader := CopyHeader(h)
	extra, err := ExtractHotstuffExtra(newHeader)
	if err != nil {
		return nil
	}

	if !keepSeal {
		extra.LeaderSeal = []byte{}
	}
	extra.AggregatedValidatorsSeal = []byte{}
	//extra.Salt = []byte{}

	payload, err := rlp.EncodeToBytes(&extra)
	if err != nil {
		return nil
	}

	newHeader.Extra = append(newHeader.Extra[:HotstuffExtraVanity], payload...)

	return newHeader
}

func HotstuffHeaderFillWithValidators(header *Header, vals []common.Address) error {
	var buf bytes.Buffer

	// compensate the lack bytes if header.Extra is not enough IstanbulExtraVanity bytes.
	if len(header.Extra) < HotstuffExtraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, HotstuffExtraVanity-len(header.Extra))...)
	}
	buf.Write(header.Extra[:HotstuffExtraVanity])

	if vals == nil {
		vals = []common.Address{}
	}
	ist := &HotstuffExtra{
		Validators:               vals,
		LeaderSeal:               []byte{},
		AggregatedValidatorsSeal: []byte{},
		Salt:                     []byte{},
	}

	payload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		return err
	}
	header.Extra = append(buf.Bytes(), payload...)
	return nil
}
