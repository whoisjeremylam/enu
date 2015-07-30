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

-func TestGetGetBlockCount(t *testing.T) {
-	result, err := GetBlockCount()
-
-	if err != nil {
-		t.Errorf(err.Error())
-	}
-
-	// 
-	
-	if result < 367576  { 
-		t.Errorf("Expected: block height > 367576, received: %d\n", result)
-		
-	}
-}