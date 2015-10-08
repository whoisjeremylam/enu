package consts

const SqlNotFound = "sql: no rows in result set"

const CounterpartylibInsufficientFunds = "insufficient funds"
const CounterpartylibMalformedAddress = "Odd-length string"
const CounterpartylibInsufficientBTC = "Insufficient BTC at address"
const CounterpartylibOnlyIssuerCanPayDividends = "only issuer can pay dividends"
const CountpartylibNoSuchAsset = "no such asset"

type ErrCodes struct {
	Code        int64
	Description string
}

type CounterpartyStruct struct {
	MiscError                 ErrCodes
	Timeout                   ErrCodes
	ReparsingOrUnavailable    ErrCodes
	SigningError              ErrCodes
	BroadcastError            ErrCodes
	InsufficientFunds         ErrCodes
	InsufficientFees          ErrCodes
	InvalidPassphrase         ErrCodes
	DividendNotFound          ErrCodes
	ComposeError              ErrCodes
	MalformedAddress          ErrCodes
	OnlyIssuerCanPayDividends ErrCodes
	NoSuchAsset               ErrCodes
}

var CounterpartyErrors = CounterpartyStruct{
	MiscError:                 ErrCodes{1000, "Misc error when contacting Counterparty. Please contact Vennd.io support."},
	Timeout:                   ErrCodes{1000, "Timeout when contacting Counterparty. Please try again later."},
	ReparsingOrUnavailable:    ErrCodes{1002, "Counterparty Blockchain temporarily unavailable. Please try again later."},
	SigningError:              ErrCodes{1003, "Unable to sign transaction. Is your passphrase correct?"},
	BroadcastError:            ErrCodes{1004, "Unable to broadcast transaction to the blockchain. Please try the transaction again."},
	InsufficientFees:          ErrCodes{1005, "Insufficient BTC in address to perform transaction. Please use the Activate() call to add more BTC."},
	InvalidPassphrase:         ErrCodes{1006, "The passphrase provided is not valid."},
	DividendNotFound:          ErrCodes{1007, "The dividend could not be found."},
	ComposeError:              ErrCodes{1008, "Unable to create the blockchain transaction."},
	InsufficientFunds:         ErrCodes{1009, "Insufficient asset in this address."},
	MalformedAddress:          ErrCodes{1010, "One of the addresses provided was not correct. Please check the addresses involved in the transaction."},
	OnlyIssuerCanPayDividends: ErrCodes{1011, "Only the issuer may pay dividends."},
	NoSuchAsset:               ErrCodes{1012, "The asset specified is incorrect or doesn't exist."},
}

type GenericStruct struct {
	InvalidDocument       ErrCodes
	InvalidDividendId     ErrCodes
	UnsupportedBlockchain ErrCodes
	HeadersIncorrect      ErrCodes
	UnknownAccessKey      ErrCodes
	InvalidSignature      ErrCodes
	InvalidNonce          ErrCodes
	NotFound              ErrCodes

	GeneralError ErrCodes
}

var GenericErrors = GenericStruct{
	InvalidDocument:       ErrCodes{1, "There was a problem with the parameters in your JSON request. Please correct the request."},
	InvalidDividendId:     ErrCodes{2, "The specified dividend id is invalid."},
	UnsupportedBlockchain: ErrCodes{3, "The specified blockchain is not supported."},
	HeadersIncorrect:      ErrCodes{4, "Request headers were not set correctly, ensure the following headers are set: accessKey and signature."},
	UnknownAccessKey:      ErrCodes{5, "Attempt to access API with unknown user key"},
	InvalidSignature:      ErrCodes{6, "Could not verify HMAC signature"},
	InvalidNonce:          ErrCodes{7, "Invalid nonce"},
	NotFound:              ErrCodes{8, "Not found"},

	GeneralError: ErrCodes{13, "Misc error. Please contact Vennd.io support."},
}
