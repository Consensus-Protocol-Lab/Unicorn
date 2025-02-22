package config

import "github.com/ethereum/go-ethereum/consensus/hotstuff/interfaces"

type Config struct {
	RequestTimeout uint64                          `toml:",omitempty"` // The timeout for each Istanbul round in milliseconds.
	BlockPeriod    uint64                          `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second for basic hotstuff and mill-seconds for event-driven
	LeaderPolicy   interfaces.SelectProposerPolicy `toml:",omitempty"` // The policy for speaker selection
	Test           bool                            `toml:",omitempty"`
	Epoch          uint64                          `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
}

// todo: modify request timeout, and miner recommit default value is 3s. recommit time should be > blockPeriod
var DefaultBasicConfig = &Config{
	RequestTimeout: 6000,
	BlockPeriod:    3,
	LeaderPolicy:   interfaces.RoundRobin,
	Epoch:          30000,
	Test:           false,
}

var DefaultEventDrivenConfig = &Config{
	RequestTimeout: 4000,
	BlockPeriod:    2000,
	LeaderPolicy:   interfaces.RoundRobin,
	Epoch:          0,
	Test:           false,
}
