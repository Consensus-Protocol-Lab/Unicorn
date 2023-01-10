package core

import (
	"github.com/ethereum/go-ethereum/log"
	"sync"
)

type Core struct {
	logger log.Logger
}

var once sync.Once
