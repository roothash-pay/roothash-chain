package kvstore

import (
	"github.com/ethereum/go-ethereum/common"
)

type PreimageSource func(key common.Hash) ([]byte, error)

type PreimageSourceSplitter struct {
	local  PreimageSource
	global PreimageSource
}

func NewPreimageSourceSplitter(local PreimageSource, global PreimageSource) *PreimageSourceSplitter {
	return &PreimageSourceSplitter{
		local:  local,
		global: global,
	}
}

func (s *PreimageSourceSplitter) Get(key [32]byte) ([]byte, error) {
	return s.global(key)
}
