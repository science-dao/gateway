package peer

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/science-dao/gateway/internal/config"
	"github.com/science-dao/gateway/internal/types"
	"github.com/science-dao/gateway/pkg/p2p"
	log "github.com/sirupsen/logrus"
)

func AuthorizeMemberPeer(peerMu ma.Multiaddr, peerId peer.ID, msg types.MemberMsg) error {
	hash := crypto.Keccak256Hash([]byte(msg.Cid))

	pk, err := crypto.SigToPub(hash.Bytes(), msg.Sig)
	if err != nil {
		log.Error(err)
		return err
	}

	err = VerifySig(pk, msg.Sig, hash.Bytes())
	if err != nil {
		return err
	}

	err = VerifyAddress(pk)
	if err != nil {
		return err
	}

	info, err := p2p.PeerMultiaddressToAddrInfo(peerMu, peerId)
	if err != nil {
		return err
	}

	n := config.Host

	p2p.AddVerifiedPeer(*n, *info)
	return nil
}

func AuthorizeBuyerPeer(peerMu ma.Multiaddr, peerId peer.ID) {
	info, err := p2p.PeerMultiaddressToAddrInfo(peerMu, peerId)
	if err != nil {
		log.Error(err)
	}

	n := config.Host

	p2p.AddVerifiedPeer(*n, *info)
}

func IsAuthorizedPeer(peerId peer.ID) bool {
	n := config.Host

	peers := p2p.GetVerifiedPeers(*n)
	for i := 0; i < len(peers); i++ {
		if peerId == peers[i] {
			return true
		}
	}
	return false
}

func VerifySig(pk *ecdsa.PublicKey, sigBytes []byte, hash []byte) error {
	publicKeyBytes := crypto.FromECDSAPub(pk)
	signatureNoRecoverID := sigBytes[:len(sigBytes)-1]
	isOk := crypto.VerifySignature(publicKeyBytes, hash, signatureNoRecoverID)
	if !isOk {
		return errors.New("the signature provided by the buyer is invalid")
	}
	return nil
}

func VerifyAddress(pk *ecdsa.PublicKey) error {
	addr := crypto.PubkeyToAddress(*pk)
	log.Debug(addr.Hex())
	// TODO: get SBT from dao -> check that address is a member

	return nil
}
