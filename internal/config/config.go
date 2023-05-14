package config

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/science-dao/gateway/internal/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var DaoConfig types.Config

var Client *ethclient.Client
var Auth *bind.TransactOpts

var Host *host.Host

func LoadConfig() {
	viper.AddConfigPath("../../")
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Failed to read env file: ", err)
	}

	err = viper.Unmarshal(&DaoConfig)
	if err != nil {
		log.Error("Failed to decode env file: ", err)
	}
}

func ConfigLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	switch DaoConfig.Mode {
	case "develop":
		log.SetLevel(log.DebugLevel)
	case "production":
		log.SetLevel(log.InfoLevel)
	}
}

func LoadClient() {
	client, err := ethclient.Dial(DaoConfig.NodeURL)
	if err != nil {
		log.Fatal(err)
	}
	Client = client
}

func LoadAuth() {
	auth, err := GetAuth(Client)
	if err != nil {
		log.Fatal("Failed to get auth object: ", err)
	}
	Auth = auth
}
