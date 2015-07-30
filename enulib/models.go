// models.go
package enulib

type Block struct {
	BlockId  int64  `json:"blockId"`
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
}

type Blocks []Block

type PaymentId struct {
	PaymentId string `json:"paymentId"`
}

type Payment struct {
	BlockId            int64  `json:"blockId"`
	SourceTxId         string `json:"sourceTxId"`
	SourceAddress      string `json:"sourceAddress"`
	DestinationAddress string `json:"destinationAddress"`
	OutAsset           string `json:"outAssest"`
	OutAmount          int64  `json:"outAmount"`
	Status             string `json:"status"`
	LastUpdatedBlockId int64  `json:"lastUpdatedblockId"`
}

type Payments []Payment

type SimplePayment struct {
	SourceAddress      string `json:"sourceAddress"`
	DestinationAddress string `json:"destinationAddress"`
	Asset              string `json:"asset"`
	Amount             uint64 `json:"amount"`
	PaymentId          string `json:"paymentId"`
	TxFee              int64  `json:"txFee"`
	BroadcastTxId      string `json:"broadcastTxId"`
	PaymentTag         string `json:"paymentTag"`
	Status             string `json:"status"`
	ErrorMessage       string `json:"errorMessage`
}

type Address struct {
	Value      string `json:"value"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

type ReturnCode struct {
	RequestId   string `json:"requestId"`
	Code        int64  `json:"code"`
	Description string `json:"description"`
}