package consts

type key int

const Satoshi = 100000000 // Default divisibility for Counterparty assets which are divisible

const RequestIdKey key = 0
const AccessKeyKey key = 1
const NonceIntKey key = 2
const BlockchainIdKey key = 3
const RequestTypeKey key = 4

const CounterpartyBlockchainId string = "counterparty"
const RippleBlockchainId string = "ripple"
const ColoredCoinsBlockchainId string = "coloredcoins"

var SupportedBlockchains = []string{CounterpartyBlockchainId, RippleBlockchainId, ColoredCoinsBlockchainId}

const AccessKeyValidStatus = "valid"       // normal status
const AccessKeyInvalidStatus = "invalid"   // the access key has been made revoked and can no longer be used
const AccessKeyDisabledStatus = "disabled" // the access key has been disabled - eg temporarily made unavailable. This can be used when maintenance is occuring on the Enu application

var AccessKeyStatuses = []string{AccessKeyValidStatus, AccessKeyInvalidStatus, AccessKeyDisabledStatus}

const SourceFile = "logger.go"
