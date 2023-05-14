package p2p

import (
	"bufio"
	"crypto/rand"
	"encoding/pem"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	log "github.com/sirupsen/logrus"
)

func GenP2PKey() (*crypto.PrivKey, error) {
	privK := open()
	if privK != nil {
		return privK, nil
	}

	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	pemPriv, err := os.Create("p2ppriv.pem")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer pemPriv.Close()

	pb, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pb,
	}
	err = pem.Encode(pemPriv, pemBlock)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &priv, nil
}

func open() *crypto.PrivKey {
	privFile, err := os.Open("p2ppriv.pem")
	if os.IsNotExist(err) {
		log.Info("Private key file doesn't exist.")
		return nil
	} else if err != nil {
		log.Error("Error opening key file: ", err)
		return nil
	}
	defer privFile.Close()

	peminfo, _ := privFile.Stat()
	size := peminfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privFile)
	_, err = buffer.Read(pembytes)
	if err != nil {
		log.Error(err)
		return nil
	}

	data, _ := pem.Decode(pembytes)

	privKey, err := crypto.UnmarshalPrivateKey(data.Bytes)
	if err != nil {
		log.Error(err)
		return nil
	}
	log.Info("Private key already exists, loading it")
	return &privKey
}
