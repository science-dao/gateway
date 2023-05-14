package utils

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"os"

	"github.com/science-dao/gateway/internal/types"
	log "github.com/sirupsen/logrus"
)

func Encrypt(cid string, fname string, f *os.File) error {
	// Get the file size
	stat, err := f.Stat()
	if err != nil {
		return err
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(f).Read(bs)
	if err != nil && err != io.EOF {
		return err
	}

	pk, pub := GenerateKeyPair(256)
	hash := sha512.New()
	cipertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, bs, nil)
	if err != nil {
		return err
	}

	err = saveJson(cid, fname, cipertext, pk)
	if err != nil {
		return err
	}

	err = os.WriteFile("encrypted.bin", cipertext, 0777)
	if err != nil {
		return err
	}

	return nil
}

// GenerateKeyPair generates a new key pair
func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Error(err)
	}
	return privkey, &privkey.PublicKey
}

// PrivateKeyToBytes private key to bytes
func PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return privBytes
}

// BytesToPrivateKey bytes to private key
func BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Error(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Error(err)
	}
	return key
}

func saveJson(cid string, fname string, ciphertext []byte, pk *rsa.PrivateKey) error {
	filename := "data.json"
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}

	fBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	kMap := make(map[string]types.FileData)
	if len(fBytes) != 0 {
		err = json.Unmarshal(fBytes, &kMap)
		if err != nil {
			return err
		}
	}

	kMap[cid] = types.FileData{
		Cid: cid,
		Key: PrivateKeyToBytes(pk),
	}

	byteJson, err := json.Marshal(kMap)
	if err != nil {
		return err
	}

	err = os.WriteFile(fname, byteJson, 0644)
	if err != nil {
		return err
	}
	return nil
}
