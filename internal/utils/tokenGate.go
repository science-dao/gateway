package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/science-dao/gateway/contracts"
	"github.com/science-dao/gateway/internal/config"
)

func CheckToken(address string, tokenId int64) bool {
	token, err := contracts.NewAccessToken(common.HexToAddress(config.DaoConfig.AccessToken), config.Client)
	if err != nil {
		return false
	}

	tokenIdBInt := big.NewInt(tokenId)

	mdata, err := token.GetTokenMetadata(nil, tokenIdBInt)
	if err != nil {
		return false
	}
	if mdata.Owner == common.HexToAddress(address) {
		return true
	}
	return false
}
