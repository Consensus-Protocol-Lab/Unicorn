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

package engine

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/config"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/core"
	event2 "github.com/ethereum/go-ethereum/consensus/hotstuff/event"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/validator"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	common2 "github.com/prysmaticlabs/prysm/v3/crypto/bls/common"
	"math/big"
	"sync"
	"time"
)

// HotStuff protocol constants.
var (
	defaultDifficulty = big.NewInt(1)
	nilUncleHash      = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
	emptyNonce        = types.BlockNonce{}
	now               = time.Now
)

type HotStuffEngine struct {
	signer          *core.Signer
	validatorNo     int
	amountValidator int
	logger          log.Logger
	config          *config.Config

	sealMu sync.Mutex
	coreMu sync.RWMutex

	commitCh          chan *types.Block
	proposedBlockHash common.Hash

	core        interfaces.HotstuffCore
	coreStarted bool

	chain          consensus.ChainReader
	currentBlock   func() *types.Block
	getBlockByHash func(hash common.Hash) *types.Block

	eventMux *event.TypeMux
}

func New(privateKey *ecdsa.PrivateKey, consensusKey *common2.SecretKey, config *config.Config, db ethdb.Database) consensus.Hotstuff {
	return &HotStuffEngine{
		signer: &core.Signer{
			EthSigner: core.NewEthSigner(privateKey, db),
			BlsSigner: core.NewBlsSigner(consensusKey, db),
		},
		config: config,
	}
}

func (e *HotStuffEngine) Author(header *types.Header) (common.Address, error) {
	return e.signer.EthSigner.Recover(header)
}

func (e *HotStuffEngine) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return e.verifyHeader(chain, header, nil, seal)
}

func (e *HotStuffEngine) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))
	go func() {
		for i, header := range headers {
			seal := false
			if seals != nil && len(seals) > i {
				seal = seals[i]
			}
			err := e.verifyHeader(chain, header, headers[:i], seal)

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// // VerifyUncles verifies that the given block's uncles conform to the consensus
// // rules of a given engine.
func (e *HotStuffEngine) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errInvalidUncleHash
	}
	return nil
}

func (e *HotStuffEngine) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	// unused fields, force to set to empty
	header.Coinbase = e.signer.EthSigner.Address()
	header.Nonce = emptyNonce
	header.MixDigest = types.HotstuffDigest

	parent, err := e.getPendingParentHeader(chain, header)
	if err != nil {
		return err
	}

	// use the same difficulty for all blocks
	header.Difficulty = defaultDifficulty

	// set header's timestamp
	header.Time = parent.Time + e.config.BlockPeriod
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}

	return nil
}

func (e *HotStuffEngine) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header) {
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = nilUncleHash
}

func (e *HotStuffEngine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// currently no rewards in hotstuff engine
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = nilUncleHash

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

func (e *HotStuffEngine) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	// update the block header timestamp and signature and propose the block to core engine
	header := block.Header()

	// sign the sig hash and fill extra seal
	if err := e.signer.EthSigner.SealBeforeCommit(header); err != nil {
		return err
	}
	block = block.WithSeal(header)

	go func() {
		// get the proposed block hash and clear it if the seal() is completed.
		e.sealMu.Lock()
		e.proposedBlockHash = block.Hash()
		e.logger.Trace("WorkerSealNewBlock", "hash", block.Hash(), "number", block.Number())

		defer func() {
			e.proposedBlockHash = common.Hash{}
			e.sealMu.Unlock()
		}()

		// post block into Hotstuff engine
		go e.EventMux().Post(event2.RequestEvent{
			Proposal: block,
		})
		for {
			select {
			case result := <-e.commitCh:
				// if the block hash and the hash from channel are the same,
				// return the result. Otherwise, keep waiting the next hash.
				if result != nil && block.Hash() == result.Hash() {
					results <- result
					return
				}
			case <-stop:
				e.logger.Trace("Stop seal, check miner status!")
				results <- nil
				return
			}
		}
	}()
	return nil
}

func (e *HotStuffEngine) SealHash(header *types.Header) common.Hash {
	return e.signer.EthSigner.SigHash(header)
}

func (e *HotStuffEngine) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return new(big.Int)
}

func (e *HotStuffEngine) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "hotstuff",
		Version:   "1.0",
		Service:   &API{chain: chain, hotstuff: e},
		Public:    true,
	}}
}

func (e *HotStuffEngine) Close() error {
	return nil
}

func (e *HotStuffEngine) Start(chain consensus.ChainReader, currentBlock func() *types.Block, getBlockByHash func(hash common.Hash) *types.Block) error {
	e.coreMu.Lock()
	defer e.coreMu.Unlock()

	if e.coreStarted {
		return ErrStartedEngine
	}

	// clear previous data
	if e.commitCh != nil {
		close(e.commitCh)
	}
	e.commitCh = make(chan *types.Block, 1)

	e.chain = chain
	e.currentBlock = currentBlock
	e.getBlockByHash = getBlockByHash

	if err := e.core.Start(chain); err != nil {
		return err
	}

	e.coreStarted = true
	return nil
}

// Stop stops the engine
func (e *HotStuffEngine) Stop() error {
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (e *HotStuffEngine) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, seal bool) error {
	if header.Number == nil {
		return errUnknownBlock
	}

	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != types.HotstuffDigest {
		return errInvalidMixDigest
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in Istanbul
	if header.UncleHash != nilUncleHash {
		return errInvalidUncleHash
	}
	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if header.Difficulty == nil || header.Difficulty.Cmp(defaultDifficulty) != 0 {
		return errInvalidDifficulty
	}

	// verifyCascadingFields verifies all the header fields that are not standalone,
	// rather depend on a batch of previous headers. The caller may optionally pass
	// in a batch of parents (ascending order) to avoid looking those up from the
	// database. This is useful for concurrently verifying a batch of new headers.
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}

	// Ensure that the block's timestamp isn't too close to it's parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if header.Time > parent.Time+e.config.BlockPeriod && header.Time > uint64(now().Unix()) {
		return errInvalidTimestamp
	}
	// Hotstuff ToDo: validator management
	vals, err := e.signer.GetValidators(e.amountValidator)
	if err != nil {
		return err
	}
	// Hotstuff ToDo: validator store
	return e.signer.VerifyHeader(header, validator.NewSet(vals, interfaces.RoundRobin), seal)
}

func (e *HotStuffEngine) getPendingParentHeader(chain consensus.ChainHeaderReader, header *types.Header) (*types.Header, error) {
	number := header.Number.Uint64()
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return nil, consensus.ErrUnknownAncestor
	}
	return parent, nil
}
