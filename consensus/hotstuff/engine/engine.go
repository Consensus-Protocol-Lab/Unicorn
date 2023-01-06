package engine

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/hotstuff"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	blst "github.com/supranational/blst/bindings/go"
	"math/big"
	"sync"
)

type Engine struct {
	address   common.Address
	ethSigner *core.EthSigner
	blsSigner *core.BlsSigner
	logger    log.Logger

	sealMu            sync.Mutex
	commitCh          chan *types.Block
	proposedBlockHash common.Hash

	eventMux *event.TypeMux
}

func New(privateKey *ecdsa.PrivateKey, consensusKey *blst.SecretKey) consensus.Hotstuff {

	blsSigner := core.NewBlsSigner(consensusKey)
	engine := &Engine{
		address:   crypto.PubkeyToAddress(privateKey.PublicKey),
		ethSigner: core.NewEthSigner(privateKey),
		blsSigner: blsSigner,
	}
	return engine
}

func (e *Engine) Author(header *types.Header) (common.Address, error) {
	return e.address, nil
}

func (e *Engine) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return nil
}

func (e *Engine) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	return nil, nil
}

//// VerifyUncles verifies that the given block's uncles conform to the consensus
//// rules of a given engine.
func (e *Engine) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return nil
}

func (e *Engine) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return nil
}

func (e *Engine) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header) {

}

func (e *Engine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	return nil, nil
}

func (e *Engine) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	// update the block header timestamp and signature and propose the block to core engine
	header := block.Header()

	// sign the sig hash and fill extra seal
	if err := e.ethSigner.SealBeforeCommit(header); err != nil {
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

		// post block into Istanbul engine
		go e.EventMux().Post(hotstuff.RequestEvent{
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

func (e *Engine) SealHash(header *types.Header) common.Hash {
	return common.HexToHash("")
}

func (e *Engine) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return nil
}

//
func (e *Engine) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "hotstuff",
		Version:   "1.0",
		Service:   &API{chain: chain, hotstuff: e},
		Public:    true,
	}}
}

//
func (e *Engine) Close() error {
	return nil
}

func (e *Engine) Start(chain consensus.ChainReader, currentBlock func() *types.Block) error {

	return nil
}

// Stop stops the engine
func (e *Engine) Stop() error {
	return nil
}

// EventMux implements hotstuff.Backend.EventMux
func (e *Engine) EventMux() *event.TypeMux {
	return e.eventMux
}
