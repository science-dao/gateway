package filecoin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/science-dao/gateway/contracts"
	"github.com/science-dao/gateway/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/web3-storage/go-w3s-client"
)

type CarResp struct {
	Size      int
	PieceCid  string
	PieceSize int
	DataCid   string
}

type CarRespBytes struct {
	Resp []byte
}

func (cResp *CarRespBytes) Write(p []byte) (n int, err error) {
	cResp.Resp = append(cResp.Resp, p...)
	return os.Stdout.Write(p)
}

func Prepare() (string, CarResp, error) {
	var resp CarRespBytes
	fpath := "./generate-car"
	command := &exec.Cmd{
		Path:   fpath,
		Args:   []string{fpath, "--single", "-i", "./encrypted.bin", "-o", "./", "-p", "./"},
		Stdout: &resp,
		Stderr: os.Stderr,
	}

	if err := command.Run(); err != nil {
		return "", CarResp{}, err
	}

	var carResult CarResp

	err := json.Unmarshal(resp.Resp, &carResult)
	if err != nil {
		return "", CarResp{}, err
	}

	f, err := os.Open("encrypted.bin")
	if err != nil {
		return "", CarResp{}, err
	}

	client, err := w3s.NewClient(w3s.WithToken(config.DaoConfig.W3SKey))
	if err != nil {
		return "", CarResp{}, err
	}
	cid, err := client.Put(context.Background(), f)
	if err != nil {
		return "", CarResp{}, err
	}

	basename := path.Base("encrypted.bin")
	gatewayURL := fmt.Sprintf("https://%s.ipfs.dweb.link/%s\n", cid.String(), basename)

	return gatewayURL, carResult, nil
}

func AddToContract(pieceCid []byte, pieceSize uint64, dataCid string, link string, carSize uint64) error {
	client, err := contracts.NewDealClient(common.HexToAddress(config.DaoConfig.DaoAddress), config.Client)
	if err != nil {
		return err
	}

	extra := contracts.StructsExtraParamsV1{
		LocationRef:        link,
		CarSize:            carSize,
		SkipIpniAnnounce:   false,
		RemoveUnsealedCopy: false,
	}

	req := contracts.StructsDealRequest{
		PieceCid:             pieceCid,
		PieceSize:            pieceSize,
		VerifiedDeal:         false,
		StartEpoch:           0,
		EndEpoch:             52000,
		StoragePricePerEpoch: big.NewInt(0),
		ProviderCollateral:   big.NewInt(0),
		ClientCollateral:     big.NewInt(0),
		DataCid:              dataCid,
		ExtraParamsVersion:   1,
		ExtraParams:          extra,
	}
	tx, err := client.MakeDealProposal(config.UpdatedAuth(), req)
	if err != nil {
		return err
	}
	log.Debug("TX sent: ", tx.Hash())
	return nil

}
