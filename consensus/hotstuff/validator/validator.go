package validator

import "github.com/ethereum/go-ethereum/common"

type Validator struct {
	address common.Address
}

type ValidatorSet struct {
	set []Validator
}
