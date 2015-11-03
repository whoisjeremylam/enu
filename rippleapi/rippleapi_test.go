package rippleapi

import (
	"testing"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"log"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var account string = "rE1Lec75PEmeDFwAuumto2Nbo8ZwG3aT9V"
var secret string = "sn2KQ6kd9NgiS1bf27j5M86U1Yyom"

var account2 string = "rzNW2QzW3S4FoQgQ7TRCks3Mpty4ADULQ"
var secret2 string = "ss86gJoUhZkZCyaiYCeY8LiMcixzK"

var accountExisting string = "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"
var falseAccount string = "SBrf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"

var destinationAccount string = "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX"
var invalidSecret string = "sn3nxiW7v8KXzPzAqzyHXbSSKNuN9"
var client_resource_id string = "4e49ef64-4729-49ce-b907-2d49ea37ac6ek"

/*
func TestHttpGet(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	_, _, err := httpGet(c, "http://localhost:5990/v1/accounts/rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn/settings")

	if err != nil {
		t.Errorf(err.Error())
	}
	log.Printf("Result ok")

}
*/

/*
func TestCreateWallet(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := CreateWallet(c)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/
/*
func TestGetAccountSettings(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	println("Account:")
	println(account)

	result, err := GetAccountSettings(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

	// negative test
	//	result, err = GetAccountSettings(c, falseAccount)

	//	if err == nil {
	//		t.Errorf("No error reported on incorrect account")
	//	}

	//	println("Result:")
	//	println(string(err.Error()))

}
*/
/*
func TestGetAccountBalances(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetAccountBalances(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/
/*
func TestPreparePayment(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := PreparePayment(c, account, account2, 1, "XRP", account2)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/
/*
func TestPostPayment(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	var issuer string = ""
	// if currency is XRP then the issuer is blank, otherwise its populated with the sender account

	// positive test
	result, err := PostPayment(c, secret, client_resource_id, account, account2, 5, "XRP", issuer)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

	// negative test
	//	_, err = PostPayment(c, invalidSecret, client_resource_id ,account2, destinationAccount, 1, "USD", account)

	//	if err == nil {
	//		t.Errorf("Error was expected, but none received.")
	//	}

	//	log.Printf ("Result: %s", result)
	//  log.Printf("Result: %s", string(err.Error()))

}
*/
/*
func TestGetConfirmPayment(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetConfirmPayment(c, account, client_resource_id)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf ("Result: %s", result)

}

*/
/*
func TestGetServerStatus(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetServerStatus(c)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/
/*
func TestPostTrustline(t *testing.T) {

	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	var limit int64 = 100
	var currency string = "SBX"

	// positive test
	result, err := PostTrustline(c, secret2, account, account2, limit, currency)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

	// negative test
	//	_, err = PostTrustline(c, invalidSecret, accountExisting, destinationAccount)

	//	if err == nil {
	//		t.Errorf("Error was expected, but none received.")
	//		t.Errorf(err.Error())
	//	}

	//	log.Printf ("Result: %s", result)
	//	log.Printf ("Result: %s", string(err.Error()))

}
*/
/*
func TestGetTrustLines(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetTrustLines(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/
/*
func TestPostAccountlines(t *testing.T) {

	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	// positive test
	result, _, err := PostAccountlines(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/
/*
func TestPostServerInfo(t *testing.T) {

	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	// positive test
	result, _, err := PostServerInfo(c)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
*/

func TestGetCurrenciesByAccount(t *testing.T) {

	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	// positive test
	result, _, err := GetCurrenciesByAccount(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf("Result: %s", result)

}
