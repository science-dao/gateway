package p2p

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
)

const (
	MemberMsgProtocol = "/dao/member/0.0.1"
	BuyerMsgProtocol  = "/dao/buyer/0.0.1"
	TransferProtocol  = "/dao/transfer/0.0.1"
)

type PendingTransfer struct {
	Cid   string
	Fname string
}

type PendingBuyerTransfer struct {
	TokenId int64
	Sig     []byte
}

var PendingTransfers map[string]PendingTransfer
var PendingBuyerTransfers map[string]PendingBuyerTransfer

func Init(port int) (*host.Host, error) {
	priv, err := GenP2PKey()
	if err != nil {
		log.Error("Error: ", err)
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.Identity(*priv),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic/", port)),
		libp2p.DefaultTransports,
		libp2p.DisableRelay(),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	PendingTransfers = make(map[string]PendingTransfer)
	PendingBuyerTransfers = make(map[string]PendingBuyerTransfer)

	return &h, nil
}

func SignPeerAddr(k string, addr string) ([]byte, error) {
	priv, err := crypto.HexToECDSA(k)
	if err != nil {
		return []byte{}, err
	}
	hash := crypto.Keccak256Hash([]byte(addr))
	sig, err := crypto.Sign(hash.Bytes(), priv)
	if err != nil {
		return []byte{}, err
	}
	return sig, nil
}

func GetVerifiedPeers(n host.Host) peer.IDSlice {
	return n.Peerstore().Peers()
}

func AddVerifiedPeer(n host.Host, info peer.AddrInfo) {
	n.Peerstore().AddAddr(info.ID, info.Addrs[0], peerstore.PermanentAddrTTL)
}

func StartListening(n host.Host, transferHandle func(s network.Stream) error, memberHandle func(s network.Stream) error, buyerHandle func(s network.Stream) error) {
	fullAddr := getHostAddress(n)
	log.Info("Node address: ", fullAddr)

	n.SetStreamHandler(MemberMsgProtocol, func(s network.Stream) {
		log.Info("Member msg received")
		if err := memberHandle(s); err != nil {
			log.Error(err)
			s.Reset()
		}
		s.Close()
	})

	n.SetStreamHandler(BuyerMsgProtocol, func(s network.Stream) {
		log.Info("Buyer msg received")
		if err := buyerHandle(s); err != nil {
			log.Error(err)
			s.Reset()
		}
		s.Close()
	})

	n.SetStreamHandler(TransferProtocol, func(s network.Stream) {
		log.Info("Trasnfer file msg received")
		if err := transferHandle(s); err != nil {
			log.Error(err)
			s.Reset()
		}
		s.Close()
	})

	log.Info("Gateway ready to receive P2P connections")
}

func PeerMultiaddressToAddrInfo(peerMu ma.Multiaddr, peerId peer.ID) (*peer.AddrInfo, error) {
	mu, err := ma.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", peerMu.String(), peerId.String()))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info(mu, "  multi")
	info, err := peer.AddrInfoFromP2pAddr(mu)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return info, nil
}

func TransferFile(n host.Host, filename string, target ma.Multiaddr) error {
	n.SetStreamHandler(TransferProtocol, func(s network.Stream) {
		log.Info("Received response")
		if err := handleFileRespMsg(s); err != nil {
			log.Error(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	info, err := peer.AddrInfoFromP2pAddr(target)
	if err != nil {
		return err
	}

	n.Peerstore().AddAddr(info.ID, info.Addrs[0], peerstore.PermanentAddrTTL)

	s, err := n.NewStream(context.Background(), info.ID, TransferProtocol)

	if err != nil {
		return err
	}

	log.Info("Opening stream to send file")

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fstat, err := file.Stat()
	if err != nil {
		return err
	}
	fileLenStr := strconv.FormatInt(fstat.Size(), 10)
	_, err = s.Write([]byte(fmt.Sprintf("%s\n", fileLenStr)))
	if err != nil {
		return err
	}

	bytesWritten, err := io.Copy(s, file)
	if err != nil {
		return err
	}
	log.Info("Sent bytes: ", bytesWritten)

	buf := bufio.NewReader(s)
	out, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.Contains(out, "FILE_RECEIVED") {
		return errors.New("buyer has not confirmed file receipt")
	}
	return nil
}

func handleFileRespMsg(s network.Stream) error {
	buf := bufio.NewReader(s)
	resp, err := buf.ReadString('\n')
	if err != nil {
		return err
	}
	log.Debug("Resp: ", resp)
	_, err = s.Write([]byte("ok"))
	return err
}

func getHostAddress(ha host.Host) string {
	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", ha.ID().Pretty()))
	if err != nil {
		log.Error("Error opening addr: ", err)
		return ""
	}
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}
