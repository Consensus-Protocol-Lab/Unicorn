package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"
	"github.com/ethereum/go-ethereum/core/types"
)

type Signer struct {
	EthSigner *EthSigner
	BlsSigner *BlsSigner
}

func (s *Signer) VerifyHeader(header *types.Header, valSet interfaces.ValidatorSet, seal bool) error {
	res := s.EthSigner.VerifyLeaderSeal(header)
	if res == nil && seal {
		res = s.BlsSigner.VerifyValidatorSeal(header, valSet)
	}
	return res
}

func (s *Signer) GetValidators(valNum int) ([]common.Address, error) {
	vals := make([]common.Address, valNum)
	for i := 0; i < valNum; i++ {
		if val, err := s.EthSigner.GetValidator(i); err != nil {
			return nil, err
		} else {
			vals[i] = val
		}
	}
	return vals, nil
}
