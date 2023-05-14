package types

type MemberMsg struct {
	Cid   string `json:"cid"`
	Sig   []byte `json:"sig"`
	Fname string `json:"file"`
}

type FileData struct {
	Cid string `json:"cid"`
	Key []byte `json:"key"`
}

type BuyerMsg struct {
	TokenId int64  `json:"tokenId"`
	Sig     []byte `json:"sig"`
}
