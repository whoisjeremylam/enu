package rippleapi

import (
	"testing"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"log"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var destinationAddress string = "1Bd5wrFxHYRkk4UCFttcPNMYzqJnQKfXUE"

var account string = "rfhpGzTZMZxrZyQdNJFiegxS7vVpqKxRiQ"
var secret string = "sn3SytUeY1WxkYBXrJQzoE1KNrZC8"

var account2 string = "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"
var falseAccount string = "SBrf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"
var destinationAccount string = "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX"
var invalidSecret string = "sn3nxiW7v8KXzPzAqzyHXbSSKNuN9"
var client_resource_id string = "4e49ef64-4729-49ce-b907-2d49ea37ac6eg"


/*
func TestHttpGet(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	_, _, err := httpGet(c, "/v1/accounts/rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn/settings")

	if err != nil {
		t.Errorf(err.Error())
	}

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

	log.Printf ("Result: %s", result)

}
*/


/*
func TestGetAccountSettings(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetAccountSettings(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf ("Result: %s", result)

// negative test
	result, err = GetAccountSettings(c, falseAccount)

	if err == nil {
		t.Errorf("No error reported on incorrect account")
	}

	println("Result:")
	println(string(err.Error()))



}


func TestGetAccountBalances(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetAccountBalances(c, account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf ("Result: %s", result)

}
*/

/*
func TestPreparePayment(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := PreparePayment(c, account, destinationAccount, 1, "USD", account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf ("Result: %s", result)

}
*/

/*
func TestPostPayment(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

// positive test
	result, err := PostPayment(c, secret, client_resource_id ,account, destinationAccount, 1, "USD", account)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf ("Result: %s", result)


// negative test
	_, err = PostPayment(c, invalidSecret, client_resource_id ,account2, destinationAccount, 1, "USD", account)

	if err == nil {
		t.Errorf("Error was expected, but none received.")
//		t.Errorf(err.Error())
	}

//	log.Printf ("Result: %s", result)
	log.Printf ("Result: %s", string(err.Error()))

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

func TestGetServerStatus(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetServerStatus(c)

	if err != nil {
		t.Errorf(err.Error())
	}

	log.Printf ("Result: %s", result)

}