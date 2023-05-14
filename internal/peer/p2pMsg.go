package peer

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/science-dao/gateway/internal/filecoin"
	"github.com/science-dao/gateway/internal/types"
	"github.com/science-dao/gateway/internal/utils"
	"github.com/science-dao/gateway/pkg/p2p"
	log "github.com/sirupsen/logrus"
)

func HandleMemberMsg(s network.Stream) error {
	// add pending transfer
	buf := bufio.NewReader(s)

	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}
	// unmarshaling into JSON
	jsonBytes := []byte(str)
	var msg types.MemberMsg
	json.Unmarshal(jsonBytes, &msg)

	err = AuthorizeMemberPeer(s.Conn().RemoteMultiaddr(), s.Conn().RemotePeer(), msg)
	if err != nil {
		return err
	}

	p2p.PendingTransfers[s.Conn().RemotePeer().String()] = p2p.PendingTransfer{Cid: msg.Cid, Fname: msg.Fname}

	_, err = s.Write([]byte("received \n"))
	return err
}

func HandleBuyerMsg(s network.Stream) error {
	buf := bufio.NewReader(s)

	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	// unmarshaling into JSON
	jsonBytes := []byte(str)
	var msg types.BuyerMsg
	json.Unmarshal(jsonBytes, &msg)

	AuthorizeBuyerPeer(s.Conn().RemoteMultiaddr(), s.Conn().RemotePeer())

	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("%d", msg.TokenId)))

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
	addr := crypto.PubkeyToAddress(*pk)
	isOk := utils.CheckToken(addr.Hex(), msg.TokenId)
	if !isOk {
		return errors.New("invalid caller")
	}

	return nil
}

func HandleFileReceival(s network.Stream) error {
	isAuthPeer := IsAuthorizedPeer(s.Conn().RemotePeer())
	if !isAuthPeer {
		s.Reset()
		return errors.New("request from unauthorized peer")
	}
	pendingTransfer := p2p.PendingTransfers[s.Conn().RemotePeer().String()]

	buf := bufio.NewReader(s)

	newfile, err := os.Create(pendingTransfer.Fname)
	if err != nil {
		return err
	}

	_, err = io.Copy(newfile, buf)
	if err != nil {
		return err
	}

	_, err = s.Write([]byte("FILE_RECEIVED \n"))
	if err != nil {
		return err
	}

	err = utils.Encrypt(pendingTransfer.Cid, pendingTransfer.Fname, newfile)
	if err != nil {
		return err
	}

	gatewayUrl, resp, err := filecoin.Prepare()
	if err != nil {
		return err
	}

	err = filecoin.AddToContract([]byte(resp.PieceCid), uint64(resp.PieceSize), resp.DataCid, gatewayUrl, uint64(resp.Size))

	return err
}
