package config

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

func GetAuth(client *ethclient.Client) (*bind.TransactOpts, error) {
	privKey, err := crypto.HexToECDSA(DaoConfig.Key)
	if err != nil {
		log.Error("Error deriving private key: ", err)
		return nil, err
	}

	pubKey := privKey.Public()
	publicKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("Error casting public key to ECDSA.")
		return nil, errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Error("Error getting nonce: ", err)
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(DaoConfig.ChainId))
	log.Info("Address for configured private key: ", auth.From)

	if err != nil {
		log.Error("Error creating tx signer: ", err)
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	return auth, nil
}

func UpdatedAuth() *bind.TransactOpts {
	nonce, err := Client.PendingNonceAt(context.Background(), Auth.From)
	if err != nil {
		log.Fatal("Error getting nonce: ", err)
	}

	Auth.Nonce = big.NewInt(int64(nonce))
	return Auth
}
