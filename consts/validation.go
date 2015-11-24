package consts

type Validations map[string]string

var ParameterValidations = map[string]Validations{
	"counterparty": {
		"asset":           `{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"description":{"type":"string"},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"},"divisible":{"type":"boolean"}},"required":["sourceAddress","asset","quantity","divisible"]}`,
		"dividend":        `{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"dividendAsset":{"type":"string"},"quantityPerUnit":{"type":"integer"}},"required":["sourceAddress","asset","dividendAsset","quantityPerUnit"]}`,
		"walletCreate":    `{"properties":{"numberOfAddresses":{"type":"integer"}}}`,
		"walletPayment":   `{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"destinationAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"}},"required":["sourceAddress","asset","quantity","destinationAddress"]}`,
		"simplePayment":   `{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"destinationAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"amount":{"type":"integer"},,"txFee":{"type":"integer"}},"required":["sourceAddress","destinationAddress","asset","amount"]}`,
		"activateaddress": `{"properties":{"address":{"type":"string","maxLength":34,"minLength":34},"amount":{"type":"integer"}},"required":["address","amount"]}`,
	},
	"ripple": {
		"walletPayment": `{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"destinationAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":3},"quantity":{"type":"integer"}},"required":["sourceAddress","asset","quantity","destinationAddress"]}`,
	},
}
