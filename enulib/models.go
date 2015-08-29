// models.go
package enulib

type Block struct {
	BlockId  int64  `json:"blockId"`
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
}

type Blocks []Block

type Amount struct {
	Asset    string `json:"asset"`
	Quantity uint64 `json:"quantity"`

}

type AddressAmount struct {
    Address           string  `json:"address"`
    Quantity          uint64  `json:"quantity"`
    PercentageHolding float64 `json:"percentageHolding"`
}


type AssetBalances struct {
	Asset        string          `json:"asset"`
	Locked       bool            `json:"locked"`
	Divisible    bool            `json:"divisible"`
	Divisibility uint64          `json:"divisibility"`
	Description  string          `json:"description"`
	Supply       uint64          `json:"quantity"`
	Balances     []AddressAmount `json:"balances"`
	RequestId	 string 		 `json:"requestId`		
}


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
	RequestId          string `json:"requestId"`		
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
	ErrorMessage       string `json:"errorMessage"`
	RequestId          string `json:"requestId"`	
}

type Address struct {
	Value      string `json:"value"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
	RequestId  string `json:"requestId"`	
}

type AddressBalances struct {
	Address  string        `json:"address"`
	Balances []Amount 		`json:"balances"`
	RequestId string       `json:"requestId`		
}

type Asset struct {
	Passphrase    string `json:"passphrase"`
	SourceAddress string `json:"sourceAddress"`
	AssetId       string `json:"assetId"`
	Asset         string `json:"asset"`
	Description   string `json:"description"`
	Quantity      uint64 `json:"quantity"`
	Divisible     bool   `json:"divisible"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"errorMessage"`
	RequestId     string `json:"requestId"`
}

type ReturnCode struct {
	RequestId   string `json:"requestId"`
	Code        int64  `json:"code"`
	Description string `json:"description"`
}

type Dividend struct {
	Passphrase      string `json:"passphrase"`
	SourceAddress   string `json:"sourceAddress"`
	DividendId      string `json:"dividendId"`
	Asset           string `json:"asset"`
	DividendAsset   string `json:"dividendAsset"`
	QuantityPerUnit uint64 `json:"quantityPerUnit"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"errorMessage"`
	RequestId       string `json:"requestId"`	
}
