package handlers

import (
	"context"

	"github.com/science-dao/gateway/internal/config"
	"github.com/science-dao/gateway/internal/peer"
	"github.com/science-dao/gateway/pkg/p2p"
	log "github.com/sirupsen/logrus"
)

func Receive() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n, err := p2p.Init(config.DaoConfig.Port)
	if err != nil {
		log.Fatal(err)
	}
	config.Host = n

	p2p.StartListening(*n, peer.HandleFileReceival, peer.HandleMemberMsg, peer.HandleBuyerMsg)
	<-ctx.Done()
}
