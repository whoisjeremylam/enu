package bitcoinapi

import (
	//"log"
	"testing"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var destinationAddress string = "1Bd5wrFxHYRkk4UCFttcPNMYzqJnQKfXUE"

func TestGetBalance(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetBalance(c, destinationAddress)

	if err != nil {
		t.Errorf(err.Error())
	}

	println(result)

	if result < 0 {
		t.Errorf("Balance is too small\n", result)

	}

}

func TestGetGetBlockCount(t *testing.T) {
	result, err := GetBlockCount()

	if err != nil {
		t.Errorf(err.Error())
	}

	//

	if result < 367576 {
		t.Errorf("Expected: block height > 367576, received: %d\n", result)

	}
}

func TestHttpGet(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	_, _, err := httpGet(c, "http://btc.blockr.io/api/v1/address/balance/198aMn6ZYAczwrE5NvNTUMyJ5qkfy4g3Hi")

	if err != nil {
		t.Errorf(err.Error())
	}
}
