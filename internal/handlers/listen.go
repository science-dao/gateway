package handlers

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/science-dao/gateway/internal/config"
	log "github.com/sirupsen/logrus"
)

func Listen() {
	address := common.HexToAddress(config.DaoConfig.DaoAddress)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{address},
	}

	logs := make(chan types.Log)
	sub, err := config.Client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Gateway node is listening for events")

	for {
		select {
		case err := <-sub.Err():
			log.Error("Failed to get event log: ", err)
		case vLog := <-logs:
			log.Debug(vLog)
			// TODO
		}
	}

}
