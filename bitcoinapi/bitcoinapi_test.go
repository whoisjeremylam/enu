package bitcoinapi

import (
	//"log"
	"testing"
	
)

var destinationAddress string = "13Yoi8aMj5ygY8dKx17vxsTMkHNP9SVayy"

func TestGetBalance(t *testing.T) {
	
	result, err := GetBalance(destinationAddress)

	if err != nil {
		t.Errorf(err.Error())
	}

	// 
	println(result)
	if result < 0  { 
		t.Errorf("Balance is too small\n", result)
		
	}

}
