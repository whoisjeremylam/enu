package counterpartyapi

import (
	"errors"
	"log"
	"reflect"
	"testing"

	"github.com/whoisjeremylam/enu/consts"
	"github.com/whoisjeremylam/enu/counterpartycrypto"
	"github.com/whoisjeremylam/enu/enulib"

	"github.com/whoisjeremylam/enu/internal/golang.org/x/net/context"
)

var c context.Context
var passphrase string = "attention stranger fate plain huge poetry view precious drug world try age"
var sendAddress string = "1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb"
var destinationAddress string = "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1"

// Use issuances for XBTC because we control the private key and won't be making more issuances
var getIssuancesExpectedTestData []Issuance = []Issuance{
	{4473, "9e87c48ffdbbd4bfa321de75181c05662fce6d24095ad3572defa8fd5be48452", 286437, "XBTC", 2100000000000000, 1, "1LHpjmevxx3ZWydTL5PfoSiUiuYNbkqknm", "1LHpjmevxx3ZWydTL5PfoSiUiuYNbkqknm", 0, "BTC", 500000000, 0, "valid"},
	{543558, "7778b2b3f085b82fb37c089fbf2737ee0cc2b3f39da3f0cce1dd5433b52fefa9", 309624, "XBTC", 0, 1, "1LHpjmevxx3ZWydTL5PfoSiUiuYNbkqknm", "1LHpjmevxx3ZWydTL5PfoSiUiuYNbkqknm", 0, "BTC", 0, 1, "valid"},
	{8829647, "a3af94b45f1c49557969a4932705b2bd1d80c83fc023346314471e20f29647d0", 328176, "XBTC", 0, 1, "1LHpjmevxx3ZWydTL5PfoSiUiuYNbkqknm", "1E6ifCRs2r6gb3pvZtBgsqnqbxuDHUSjU9", 1, "XBTC", 0, 0, "valid"},
}

// Use balances for MPUSD - it looks abandoned
var getBalancesByAssetExpectedTestData []Balance = []Balance{
	{Quantity: 111100000092, Asset: "MPUSD", Address: "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy"},
	{Quantity: 888899999908, Asset: "MPUSD", Address: "1Do5J89372i9bv3UQe6Uf6NG74ZW1Riwaa"},
}

// Use balances for 1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy - it looks abandoned
var getBalancesByAddressExpectedTestData []Balance = []Balance{
	{Quantity: 537758615, Asset: "MPBTC", Address: "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy"},
	{Quantity: 111100000092, Asset: "MPUSD", Address: "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy"},
	{Quantity: 250482117835, Asset: "XCP", Address: "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy"},
}

func setContext() {
	c = context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())
	c = context.WithValue(c, consts.AccessKeyKey, "unittesting")
	c = context.WithValue(c, consts.BlockchainIdKey, "counterparty")
	c = context.WithValue(c, consts.EnvKey, "dev")
}

func TestGetIssuances(t *testing.T) {
	Init()
	setContext()

	resultGetIssuances, errorCode, err := GetIssuances(c, "XBTC")
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(resultGetIssuances) == 0 {
		t.Errorf("Expected: resultGetIssuances to contain value, Got: %+v errorCode=%d\n", resultGetIssuances, errorCode)
	}

	if reflect.DeepEqual(resultGetIssuances, getIssuancesExpectedTestData) != true {
		t.Errorf("Expected: %s, got %s", getIssuancesExpectedTestData, resultGetIssuances)
	}
}

func TestGetIssuancesDB(t *testing.T) {
	Init()
	setContext()

	resultGetIssuances, errorCode, err := GetIssuancesDB(c, "XBTC")

	if err != nil {
		t.Errorf(err.Error())
	}

	if len(resultGetIssuances) == 0 {
		t.Errorf("Expected: resultGetIssuances to contain value, Got: %+v errorCode=%d\n", resultGetIssuances, errorCode)
	}

	if reflect.DeepEqual(resultGetIssuances, getIssuancesExpectedTestData) != true {
		t.Errorf("Expected: %s, got %s", getIssuancesExpectedTestData, resultGetIssuances)
	}
}

func TestGenerateRandomAssetName(t *testing.T) {
	setContext()

	result, err := generateRandomAssetName(c)

	if err != nil {
		t.Errorf(err.Error())
	}

	// We can do more validation tests here if the numeric portion exceeds the range of integers allowed
	if len(result) < 18 {
		t.Errorf("The asset name that was generated is too small\n")
	}

}

func TestGetBalancesByAsset(t *testing.T) {
	Init()
	setContext()

	// Not a real asset
	resultGetBalances, errorCode, err := GetBalancesByAsset(c, "NOTASSETREALLY")
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(resultGetBalances) != 0 {
		t.Errorf("Expected: resultGetBalances to be empty value, Got: %+v status: %d\n", resultGetBalances, errorCode)
	}

	// Abandoned asset
	resultGetBalances, errorCode, err = GetBalancesByAsset(c, "MPUSD")
	if err != nil || errorCode != 0 {
		t.Errorf(err.Error())
	}

	if reflect.DeepEqual(resultGetBalances, getBalancesByAssetExpectedTestData) != true {
		t.Errorf("Expected: %#v, Got: %#v\n", getBalancesByAssetExpectedTestData, resultGetBalances)
	}
}

func TestGetBalancesByAddress(t *testing.T) {
	Init()
	setContext()

	// Retrieve address which is empty
	resultGetBalances, errorCode, err := GetBalancesByAddress(c, "1enuEmptyAdd8ALj6mfBsbifRoD4miY36v")
	if err != nil || errorCode != 0 {
		t.Errorf(err.Error())
	}
	if len(resultGetBalances) != 0 {
		t.Errorf("Expected: resultGetBalances = [], Got: %s\n", resultGetBalances)
	}

	// Retrieve 1enuEmptyAdd8ALj6mfBsbifRoD4miY36v which is abandoned
	resultGetBalances, errorCode, err = GetBalancesByAddress(c, "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy")
	if err != nil || errorCode != 0 {
		t.Errorf(err.Error())
	}

	if reflect.DeepEqual(resultGetBalances, getBalancesByAddressExpectedTestData) != true {
		t.Errorf("Expected: %#v, Got: %#v\n", getBalancesByAddressExpectedTestData, resultGetBalances)
	}

}

func TestGetSendsByAddress(t *testing.T) {
	Init()
	setContext()

	// Retrieve address which is empty
	resultGetSendsByAddress, errorCode, err := GetSendsByAddress(c, "19An2wpGDyDwES8hXNvDovc49ihc7iNMMD")
	log.Printf("%+#v", resultGetSendsByAddress)
	if err != nil || errorCode != 0 {
		t.Errorf(err.Error())
	}
	if len(resultGetSendsByAddress) == 0 {
		t.Errorf("Expected: resultGetSendsByAddress != [], Got: %s\n", resultGetSendsByAddress)
	}

	//	// Retrieve 1enuEmptyAdd8ALj6mfBsbifRoD4miY36v which is abandoned
	//	resultGetBalances, errorCode, err = GetBalancesByAddress(c, "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy")
	//	if err != nil || errorCode != 0 {
	//		t.Errorf(err.Error())
	//	}

	//	if reflect.DeepEqual(resultGetBalances, getBalancesByAddressExpectedTestData) != true {
	//		t.Errorf("Expected: %#v, Got: %#v\n", getBalancesByAddressExpectedTestData, resultGetBalances)
	//	}

}

func TestGetSendsByAddressDB(t *testing.T) {
	Init()
	setContext()

	// Retrieve address which is empty
	resultGetSendsByAddress, errorCode, err := GetSendsByAddressDB(c, "19An2wpGDyDwES8hXNvDovc49ihc7iNMMD")
	log.Printf("%+#v", resultGetSendsByAddress)
	if err != nil || errorCode != 0 {
		t.Errorf(err.Error())
	}
	if len(resultGetSendsByAddress) == 0 {
		t.Errorf("Expected: resultGetSendsByAddress != [], Got: %s\n", resultGetSendsByAddress)
	}

	//	// Retrieve 1enuEmptyAdd8ALj6mfBsbifRoD4miY36v which is abandoned
	//	resultGetBalances, errorCode, err = GetBalancesByAddress(c, "1MPUSDQ7MVrqbSTFfacNP1V9KBooz9XKgy")
	//	if err != nil || errorCode != 0 {
	//		t.Errorf(err.Error())
	//	}

	//	if reflect.DeepEqual(resultGetBalances, getBalancesByAddressExpectedTestData) != true {
	//		t.Errorf("Expected: %#v, Got: %#v\n", getBalancesByAddressExpectedTestData, resultGetBalances)
	//	}

}

func TestCreateSend(t *testing.T) {
	var testData = []struct {
		From              string
		To                string
		Asset             string
		Amount            uint64
		ExpectedResult    string
		ExpectedErrorCode int64
		CaseDescription   string
	}{
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 1000, "01000000027bdd9ae0e62a2ca7e5cbf3af06996632e63d52941ca4f6b9f3d1f9ca04e6f431000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff366878895666dbc4f343bf5fe05970026eeaccbb9082b49f3ddd50c35f2ef48c010000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1c9715891cbbae121b002b359e2d7792c7f3196a3ced96d35ad728756ef45e0700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", 0, "Successful case"},
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 9999, "", consts.CounterpartyErrors.InsufficientFunds.Code, "Insufficient counterparty token"},
		{"1Badaddress", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 1000, "", consts.CounterpartyErrors.MalformedAddress.Code, "Address cannot be derived from passphrase"},
	}

	Init()
	setContext()

	for _, s := range testData {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, s.From)

		if err != nil && err.Error() != "Private and public keys not found for address: 1Badaddress" {
			t.Error(err.Error())
		}

		resultCreateSend, errorCode, err := CreateSend(c, s.From, s.To, "SHIMA", s.Amount, pubKey)

		if s.ExpectedResult != resultCreateSend || s.ExpectedErrorCode != errorCode {
			if err == nil {
				err = errors.New("No error")
			}
			t.Errorf("Expected: %s errorCode=%d, Got: %s errorCode=%d\nCase: %s\n%s\nRequestId:%s\n", s.ExpectedResult, s.ExpectedErrorCode, resultCreateSend, errorCode, s.CaseDescription, err.Error(), c.Value(consts.RequestIdKey))
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
		{passphrase, "010000000241b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1c7b4ca6c1f0494d0c3130c48853dfdde4d7c5bc46552b723f7a9500a4107a0700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "010000000241b2f1a5acdf198dbb6d1f79c1f64b9bc75589ef0449f8cc6219e63af24de4c7000000006a47304402203208c8ce67d3a7ec0c06e725f6f336da292e120b79ba45eeda7aacacc02a3f51022044a975f4bcdd7ffcc9ccebdedc5dc23cd01b63df0dcddeb3b30e0121fcb589760121026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb6ffffffff09be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741000000006a473044022000886f6c52ff6740f0fdb570d2db55f71be0ebdf52b997bb06899e6985e898d30220453fd04c82a6df614743bcb7cf87ed3b72bc92931347fbab043d8fda043c12f30121026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb6ffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1c7b4ca6c1f0494d0c3130c48853dfdde4d7c5bc46552b723f7a9500a4107a0700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", "Successful test tx with 2 txins"},
		{passphrase, "010000000109be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741010000001976a914b676d3212ba3532d234b1b09f21c83d437b9507088acffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1ce940a04c56a5340496081336e4b77e9a9ee153672a7a49a86681a953ba1f4400000000001976a914b676d3212ba3532d234b1b09f21c83d437b9507088ac00000000", "010000000109be6bb4793c357111b3915a79419c5a789c82002509322f29b0f210f8c8b741010000006a473044022021cb8ed89a682fcc49482ef3e60fe560fc402e6270019afcbab662079d7e5e3c0220179547b95af8d87b81dde8f3bfeee0c70ac6c7d004bf385839fa16a8a964ad4a01210375f15dbf58283272224893c533fd046b11be427885a48b120b4be9395e3cf21cffffffff0336150000000000001976a914b889eba98a2026448b6acab4a71a1d22590ddd5888ac00000000000000001e6a1ce940a04c56a5340496081336e4b77e9a9ee153672a7a49a86681a953ba1f4400000000001976a914b676d3212ba3532d234b1b09f21c83d437b9507088ac00000000", "Successful test tx with 1 txin"},
	}

	Init()
	setContext()

	for _, s := range testData {
		result, err := SignRawTransaction(c, s.Passphrase, s.UnsignedTx)

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
	setContext()

	pubKey, err := counterpartycrypto.GetPublicKey(passphrase, "1HdnKzzCKFzNEJbmYoa3RcY4MhKPP3NB7p")
	if err != nil {
		t.Error(err.Error())

		return
	}

	payload, errorCode, err := CreateSend(c, "1HdnKzzCKFzNEJbmYoa3RcY4MhKPP3NB7p", "1HpkZBjNFRFagyj6Q2adRSagkfNDERZhg1", "SHIMA", 2000, pubKey)
	if err != nil || errorCode != 0 {
		t.Error(err.Error())

		return
	}

	signedRawTransactionHexString, err := SignRawTransaction(c, "attention stranger fate plain huge poetry view precious drug world try age", payload)
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
		SourceAddress     string
		Asset             string
		Description       string
		Quantity          uint64
		Divisible         bool
		ExpectedResult    string
		ExpectedErrorCode int64
		CaseDescription   string
	}{
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "JOGHJOHV", "JOGHJOHV", 1000000, true, "01000000027bdd9ae0e62a2ca7e5cbf3af06996632e63d52941ca4f6b9f3d1f9ca04e6f431000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff366878895666dbc4f343bf5fe05970026eeaccbb9082b49f3ddd50c35f2ef48c010000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff02781e00000000000069512102fb18931ab9ac1416592b359e397792c7e2bcc3de3696d35ad72879c4d8275e8821022fd607edaea20dc24365f15c8e62515a7bf3fbe7e2868feafd10213e0bcac04021026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb653aeb2550700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", 0, "Alphabetic asset name"},
		{"1CipmbDRHn89cgqs6XbjswkrDxvCKA8Tfb", "A8133401331811274061", "TEST ASSET", 1000000, true, "01000000027bdd9ae0e62a2ca7e5cbf3af06996632e63d52941ca4f6b9f3d1f9ca04e6f431000000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff366878895666dbc4f343bf5fe05970026eeaccbb9082b49f3ddd50c35f2ef48c010000001976a9148092503d3303106c4844c639db0f60298c573f7488acffffffff02781e00000000000069512103e518931ab9ac1416592b359e39074d6298485ce1a096d35ad72879c4d8275e4421022fd607edaea20dc2417bfb4892085f417eb6afe7e2868feafd10213e0bcac00521026e2d0f2ca390f63c6e8786fa48d33544427997dbe4a9ebac14ffe8c8ef903bb653aeb2550700000000001976a9148092503d3303106c4844c639db0f60298c573f7488ac00000000", 0, "Numeric asset name"},
		{"19kXH7PdizT1mWdQAzY9H4Yyc4iTLTVT5A", "JOGHJOHV", "JOGHJOHV", 1000000, true, "", consts.CounterpartyErrors.InsufficientFunds.Code, "Alphabetic asset name from address with insufficient XCP"},
		{"19kXH7PdizT1mWdQAzY9H4Yyc4iTLTVT5A", "A8133401331811274061", "TEST ASSET", 1000000, true, "", consts.CounterpartyErrors.InsufficientFees.Code, "Numeric asset name from address with insufficient BTC"},
	}

	Init()
	setContext()
	c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())

	for _, s := range testData {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, s.SourceAddress)
		if err != nil {
			t.Error(err.Error())
		}

		if pubKey == "" {
			t.Errorf("Unable to retrieve pubkey for: %s\n", s.SourceAddress)
		}

		resultCreateIssuance, errorCode, err := createIssuance(c, s.SourceAddress, s.Asset, s.Description, s.Quantity, s.Divisible, pubKey)

		if s.ExpectedResult != resultCreateIssuance || s.ExpectedErrorCode != errorCode {
			if err == nil {
				err = errors.New("No error")
			}

			t.Errorf("Expected: %s errorCode=%d, Got: %s errorCode=%d\nCase: %s\n%s\nRequestId:%s\n", s.ExpectedResult, s.ExpectedErrorCode, resultCreateIssuance, errorCode, s.CaseDescription, err.Error(), c.Value(consts.RequestIdKey))
		}
	}
}

func TestCreateDividend(t *testing.T) {
	var testData = []struct {
		SourceAddress     string
		Asset             string
		DividendAsset     string
		QuantityPerUnit   uint64
		ExpectedResult    string
		ExpectedErrorCode int64
		CaseDescription   string
	}{
		{"18GgqxHNFeRiPe1yW1VsBdUp38WGeyBEpp", "CREATEDIVIDENDBADNAME", "XCP", 1, "", consts.CounterpartyErrors.NoSuchAsset.Code, "No such asset"},
		{"18GgqxHNFeRiPe1yW1VsBdUp38WGeyBEpp", "FLDC", "XCP", 1000, "", consts.CounterpartyErrors.OnlyIssuerCanPayDividends.Code, "Only issuer can pay dividends"},
	}

	Init()
	setContext()

	for _, s := range testData {
		pubKey, err := counterpartycrypto.GetPublicKey(passphrase, s.SourceAddress)
		if err != nil {
			t.Error(err.Error())
		}

		if pubKey == "" {
			t.Errorf("Unable to retrieve pubkey for: %s\n", s.SourceAddress)
		}

		resultCreateIssuance, errorCode, err := CreateDividend(c, s.SourceAddress, s.Asset, s.DividendAsset, s.QuantityPerUnit, pubKey)

		if s.ExpectedResult != resultCreateIssuance || s.ExpectedErrorCode != errorCode {
			if err == nil {
				err = errors.New("No error")
			}

			t.Errorf("Expected: %s errorCode=%d, Got: %s errorCode=%d\nCase: %s\n%s\nRequestId:%s\n", s.ExpectedResult, s.ExpectedErrorCode, resultCreateIssuance, errorCode, s.CaseDescription, err.Error(), c.Value(consts.RequestIdKey))
		}
	}
}

func TestGetRunningInfo(t *testing.T) {
	var result RunningInfo

	Init()
	setContext()

	result, errorCode, err := GetRunningInfo(c)
	if err != nil {
		t.Errorf("errorCode=%d, error=%s, RequestId=%s\n", errorCode, err.Error(), c.Value(consts.RequestIdKey))
	}

	if result.DbCaughtUp != true {
		t.Errorf("DbCaughtUp expected: true, Got: false\n")
	}
}

func TestCalculateFee(t *testing.T) {
	var testData = []struct {
		Env             string
		BlockchainId    string
		Amount          uint64
		ExpectedAmount  uint64
		ExpectedAsset   string
		CaseDescription string
	}{
		{"dev", consts.CounterpartyBlockchainId, 0, 138600, "BTC", "Specified 0, returns fee for 20 dev transactions"},
		{"prd", consts.CounterpartyBlockchainId, 0, 308600, "BTC", "Specified 0, returns fee for 20 prd transactions"},
		{"dev", consts.CounterpartyBlockchainId, 5607, 6930000, "BTC", "Specified 5607, returns fee for 1000 dev transactions"},
		{"prd", consts.CounterpartyBlockchainId, 64567, 15430000, "BTC", "Specified 64567, returns fee for 1000 prd transactions"},
		{"dev", consts.CounterpartyBlockchainId, 555, 3846150, "BTC", "Specified 555, returns fee for 1000 dev transactions"},
		{"prd", consts.CounterpartyBlockchainId, 444, 6850920, "BTC", "Specified 444, returns fee for 1000 prd transactions"},
		{"prd", consts.CounterpartyBlockchainId, 444, 6850920, "BTC", "Specified 444, returns fee for 1000 prd transactions"},
		{"prd", consts.RippleBlockchainId, 444, 0, "", "Specified an invalid blockchain, returns 0"},
	}

	for _, s := range testData {
		c := context.TODO()
		c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())
		c = context.WithValue(c, consts.EnvKey, s.Env)
		c = context.WithValue(c, consts.BlockchainIdKey, s.BlockchainId)

		resultAmount, resultAsset, _ := CalculateFeeAmount(c, s.Amount)

		if resultAmount != s.ExpectedAmount || resultAsset != s.ExpectedAsset {
			t.Errorf("Expected: %d %s, Got: %d %s\nCase: %s\n", s.ExpectedAmount, s.ExpectedAsset, resultAmount, resultAsset, s.CaseDescription)
		}
	}
}

func TestCalculateNumberOfTransactions(t *testing.T) {
	var testData = []struct {
		Env             string
		BlockchainId    string
		Amount          uint64
		ExpectedNumber  uint64
		CaseDescription string
	}{
		{"dev", consts.CounterpartyBlockchainId, 0, 0, "Specified 0 in dev"},
		{"prd", consts.CounterpartyBlockchainId, 0, 0, "Specified 0 in prd"},
		{"dev", consts.CounterpartyBlockchainId, 138600, 20, "20 transactions in dev"},
		{"prd", consts.CounterpartyBlockchainId, 308600, 20, "20 transactions in prd"},
		{"dev", consts.CounterpartyBlockchainId, 6930000, 1000, "1000 transactions in dev"},
		{"prd", consts.CounterpartyBlockchainId, 15430000, 1000, "1000 transactions in prd"},
		{"dev", consts.CounterpartyBlockchainId, 5930003, 855, "855 transactions in dev (amounts are trunced)"},
		{"prd", consts.CounterpartyBlockchainId, 13430002, 870, "870 transactions in prd (amounts are trunced)"},
		{"prd", consts.RippleBlockchainId, 444, 0, "Specified an invalid blockchain, returns 0"},
	}

	for _, s := range testData {
		c := context.TODO()
		c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())
		c = context.WithValue(c, consts.EnvKey, s.Env)
		c = context.WithValue(c, consts.BlockchainIdKey, s.BlockchainId)

		resultAmount, _ := CalculateNumberOfTransactions(c, s.Amount)

		if resultAmount != s.ExpectedNumber {
			t.Errorf("Expected: %d, Got: %d\nCase: %s\n", s.ExpectedNumber, resultAmount, s.CaseDescription)
		}
	}
}
