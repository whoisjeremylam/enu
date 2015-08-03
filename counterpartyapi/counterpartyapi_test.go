package counterpartyapi

import (
	//	"log"
	"testing"

	"github.com/vennd/enu/counterpartycrypto"
)

var passphrase string = "attention stranger fate plain huge poetry view precious drug world try age"
var sendAddress string = "1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb"
var destinationAddress string = "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1"

func TestGenerateRandomAssetName(t *testing.T) {
	result, err := generateRandomAssetName()

	if err != nil {
		t.Errorf(err.Error())
	}

	// We can do more validation tests here if the numeric portion exceeds the range of integers allowed
	if len(result) < 18 {
		t.Errorf("The asset name that was generated is too small\n")
	}

}

func TestGetBalances(t *testing.T) {
	Init()

	resultGetBalances, err := GetBalancesByAsset("XBTC")

	if err != nil {
		t.Errorf(err.Error())
	}

	if len(resultGetBalances) == 0 {
		t.Errorf("Expected: resultGetBalances to contain value, Got: %+v\n", resultGetBalances)
	}
}

func TestGetBalancesByAsset(t *testing.T) {
	Init()

	resultGetBalances, err := GetBalancesByAddress("1enuEmptyAdd8ALj6mfBsbifRoD4miY36v")

	if err != nil {
		t.Errorf(err.Error())
	}

	if len(resultGetBalances) != 0 {
		t.Errorf("Expected: resultGetBalances = [], Got: %s\n", resultGetBalances)
	}
}

func TestCreateSend(t *testing.T) {
	var testData = []struct {
		From           string
		To             string
		Asset          string
		Amount         uint64
		ExpectedResult string
		//		ExpectedError   string
		CaseDescription string
	}{
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 1000, "01000000034b3687a1a10d2613d7ec54a7fb8e845eb9bd75468a999402443163a85dd3c62c000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff41b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1cf4dd66ab2270b6d5fe5cd684372883bc787751725b08ac68c68358477ab00700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "Successful case"},
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 9999, "", "Insufficient counterparty token"},
		{"1Badaddress", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 1000, "", "Address cannot be derived from passphrase"},
	}

	Init()

	for _, s := range testData {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, s.From)

		if err != nil && err.Error() != "Private and public keys not found for address: 1Badaddress" {
			t.Error(err.Error())
		}

		resultCreateSend, err := CreateSend(s.From, s.To, "SHIMA", s.Amount, pubKey)

		if s.ExpectedResult != resultCreateSend {
			t.Errorf("Expected: %s, Got: %s\nCase: %s\n", s.ExpectedResult, resultCreateSend, s.CaseDescription)

			if err != nil && err.Error() != "Private and public keys not found for address: 1Badaddress" {
				t.Error(err.Error())
			}
		}
	}

}

func TestSignRawTransaction(t *testing.T) {
	var testData = []struct {
		Passphrase      string
		UnsignedTx      string
		ExpectedResult  string
		CaseDescription string
	}{
		{passphrase, "010000000241b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1c7b4ca6c1f0494d0c3130c48853dfdde4d7c5bc46552b723f7a9500a4107a0700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "010000000241b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000006a47304402203208c8ce67d3a7ec0c06e725f6f336da292e120b79ba45eeda7aacacc02a3f51022044a975f4bcdd7ffcc9ccebdedc5dc23cd01b63df0dcddeb3b30e0121fcb589760121026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb6ffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000006a47304402203208c8ce67d3a7ec0c06e725f6f336da292e120b79ba45eeda7aacacc02a3f51022044a975f4bcdd7ffcc9ccebdedc5dc23cd01b63df0dcddeb3b30e0121fcb589760121026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb6ffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1c7b4ca6c1f0494d0c3130c48853dfdde4d7c5bc46552b723f7a9500a4107a0700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "Successful test tx with 2 txins"},
		{passphrase, "010000000109be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741010000001976a914b676d3212ba3532d234b1b09f21c83d437b9507088acffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1ce940a04c56a5340496081336e4b77e9a9ee153672a7a49a86681a953ba1f4400000000001976a914b676d3212ba3532d234b1b09f21c83d437b9507088ac00000000", "010000000109be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741010000006a473044022021cb8ed89a682fcc49482ef3e60fe560fc402e6270019afcbab662079d7e5e3c0220179547b95af8d87b81dde8f3bfeee0c70ac6c7d004bf385839fa16a8a964ad4a01210375f15dbf58283272224893c533fd046b11be427885a48b120b4be9395e3cf21cffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1ce940a04c56a5340496081336e4b77e9a9ee153672a7a49a86681a953ba1f4400000000001976a914b676d3212ba3532d234b1b09f21c83d437b9507088ac00000000", "Successful test tx with 1 txin"},
	}

	Init()

	for _, s := range testData {
		result, err := SignRawTransaction(s.Passphrase, s.UnsignedTx)

		if s.ExpectedResult != result {
			t.Errorf("Expected: %s, Got: %s\n", s.ExpectedResult, result)

			if err != nil {
				t.Error(err.Error())
			}
		}
	}

}

func TestSendRawTransaction(t *testing.T) {
	Init()

	pubKey, err := counterpartycrypto.GetPublicKey(passphrase, "1HdnKzzCKFzNEJbmYoa3RcY4MhKPP3NB7p")
	if err != nil {
		t.Error(err.Error())

		return
	}

	payload, err := CreateSend("1HdnKzzCKFzNEJbmYoa3RcY4MhKPP3NB7p", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 2000, pubKey)
	if err != nil {
		t.Error(err.Error())

		return
	}

	signedRawTransactionHexString, err := SignRawTransaction("attention stranger fate plain huge poetry view precious drug world try age", payload)
	if err != nil {
		t.Error(err.Error())
	}

	if signedRawTransactionHexString == "" {
		t.Error("Nothing returned from Counterparty_SignRawTransaction()")
	}

	// Uncomment below lines and "log" import to send!!
	//	txId, err := Bitcoin_SendRawTransaction(signedRawTransactionHexString)
	//	if err != nil {
	//		t.Error(err.Error())
	//	}
	//	log.Printf("Success! TxId: %s\n", txId)
}

func TestCreateIssuance(t *testing.T) {
	var testData = []struct {
		SourceAddress  string
		Asset          string
		Description    string
		Quantity       uint64
		Divisible      bool
		ExpectedResult string
		//		ExpectedError   string
		CaseDescription string
	}{
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "JOGHJOHV", "JOGHJOHV", 1000000, true, "01000000034b3687a1a10d2613d7ec54a7fb8e845eb9bd75468a999402443163a85dd3c62c000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff41b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff02781e0000000000006951210298d07cad2072b0d8a75cd684232883bc69d2f8908008ac68c68354ed3a639c032102478c64444c18c4916693d58aa155339adfef540e1d412ad249528f36fbdc7af821026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb653ae38a70700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "Alphabetic asset name"},
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "A8133401331811274061", "TEST ASSET", 1000000, true, "01000000034b3687a1a10d2613d7ec54a7fb8e845eb9bd75468a999402443163a85dd3c62c000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff41b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff02781e0000000000006951210286d07cad2072b0d8a75cd68423585c19132667af1608ac68c68354ed3a639c8d2103478c64444c18c491648ddf9ebd3f3d81daaa000e1d412ad249528f36fbdc7a1921026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb653ae38a70700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "Numeric asset name"},
		{"19kXH7PdizT1mWdQAzY9H4Yyc4iTLTVT5A", "JOGHJOHV", "JOGHJOHV", 1000000, true, "", "Alphabetic asset name from address with insufficient XCP"},
		{"19kXH7PdizT1mWdQAzY9H4Yyc4iTLTVT5A", "A8133401331811274061", "TEST ASSET", 1000000, true, "", "Numeric asset name from address with insufficient BTC"},
	}

	Init()

	for _, s := range testData {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, s.SourceAddress)
		if err != nil {
			t.Error(err.Error())
		}

		if pubKey == "" {
			t.Errorf("Unable to retrieve pubkey for: %s\n", s.SourceAddress)
		}

		resultCreateIssuance, err := CreateIssuance(s.SourceAddress, s.Asset, s.Description, s.Quantity, s.Divisible, pubKey)

		if s.ExpectedResult != resultCreateIssuance {
			t.Errorf("Expected: %s, Got: %s\nCase: %s\n", s.ExpectedResult, resultCreateIssuance, s.CaseDescription)

			// Additionally log the error if we got an error
			if err != nil {
				t.Error(err.Error())
			}
		}

	}

}

func TestCreateDividend(t *testing.T) {
	var testData = []struct {
		SourceAddress   string
		Asset           string
		DividendAsset   string
		QuantityPerUnit uint64
		ExpectedResult  string
		CaseDescription string
	}{
		{"18GgqxHNFeRiPe1yW1VsBdUp38WGeyBEpp", "JOGHJOHV", "XCP", 1, "", "No such asset"},
		{"18GgqxHNFeRiPe1yW1VsBdUp38WGeyBEpp", "FLDC", "XCP", 1000, "", "Only issuer can pay dividends"},
	}

	Init()

	for _, s := range testData {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, s.SourceAddress)
		if err != nil {
			t.Error(err.Error())
		}

		if pubKey == "" {
			t.Errorf("Unable to retrieve pubkey for: %s\n", s.SourceAddress)
		}

		resultCreateIssuance, err := CreateDividend(s.SourceAddress, s.Asset, s.DividendAsset, s.QuantityPerUnit, pubKey)

		if s.ExpectedResult != resultCreateIssuance {
			t.Errorf("Expected: %s, Got: %s\nCase: %s\n", s.ExpectedResult, resultCreateIssuance, s.CaseDescription)

			// Additionally log the error if we got an error
			if err != nil {
				t.Error(err.Error())
			}
		}

	}

}

func TestGetRunningInfo(t *testing.T) {
	var result RunningInfo
	Init()

	result, err := GetRunningInfo()
	if err != nil {
		t.Error(err.Error())
	}

	if result.DbCaughtUp != true {
		t.Errorf("DbCaughtUp expected: true, Got: false\n")
	}
}
