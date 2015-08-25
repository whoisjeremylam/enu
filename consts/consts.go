package consts

type key int

const Satoshi = 100000000 // Default divisibility for Counterparty assets which are divisible

const RequestIdKey key = 0
const AccessKeyKey key = 1
const NonceIntKey key = 2
const BlockchainIdKey key = 2

const CounterpartyBlockchainId string = "counterparty"
const RippleBlockchainId string = "ripple"
const ColoredCoinsBlockchainId string = "coloredcoins"
