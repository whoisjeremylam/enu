// database.go
package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/vennd/enu/enulib"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB
var databaseString string
var isInit bool = false // set to true only after the init sequence is complete

// Initialises global variables and database connection for all handlers
func Init() {
	var configuration interface{}

	if isInit == true {
		return
	}

	// Read configuration from file
	file, err := ioutil.ReadFile("enuapi.json")
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Fatalln(err)
	}

	err = json.Unmarshal(file, &configuration)

	m := configuration.(map[string]interface{})

	//	localMode := m["localMode"].(string) // True if running on local Windows development machine
	dbUrl := m["dburl"].(string)         // URL for MySQL
	schema := m["schema"].(string)       // Database schema name
	user := m["dbuser"].(string)         // User name for the DB
	password := m["dbpassword"].(string) // Password for the specified database

	stringsToConcatenate := []string{user, ":", password, "@", dbUrl, "/", schema}
	databaseString = strings.Join(stringsToConcatenate, "")

	log.Printf("Opening: %s\n", strings.Join([]string{dbUrl, "/", schema}, ""))
	Db, err = sql.Open("mysql", databaseString)
	if err != nil {
		panic(err.Error())
	}

	// Ping to check DB connection is okay
	err = Db.Ping()
	if err != nil {
		panic(err.Error())
	}

	log.Println("Opened DB successfully!")

	isInit = true
}

// Inserts a payment into the payment database
func InsertPayment(accessKey string, blockIdValue int64, sourceTxidValue string, sourceAddressValue string, destinationAddressValue string, outAssetValue string, outAmountValue uint64, statusValue string, lastUpdatedBlockIdValue int64, txFeeValue uint64, paymentTag string) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into payments(accessKey, blockId, sourceTxid, sourceAddress, destinationAddress, outAsset, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?)")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(accessKey, blockIdValue, sourceTxidValue, sourceAddressValue, destinationAddressValue, outAssetValue, outAmountValue, statusValue, lastUpdatedBlockIdValue, txFeeValue, paymentTag)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
}

func GetPaymentByPaymentId(accessKey string, paymentId string) enulib.SimplePayment {
	if isInit == false {
		Init()
	}

	//	 Query DB
	stmt, err := Db.Prepare("select rowId, blockId, sourceTxId, sourceAddress, destinationAddress, outAsset, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag, errorDescription from payments where sourceTxid=? and accessKey=?")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(paymentId, accessKey)
	if err != nil {
		panic(err.Error())
	}

	var rowId string
	var blockId string
	var sourceAddress string
	var destinationAddress string
	var asset string
	var amount uint64
	var txFee int64
	var broadcastTxId string
	var status string
	var sourceTxId string
	var lastUpdatedBlockId string
	var payment enulib.SimplePayment
	var paymentTag string
	var errorMessage string

	if err := row.Scan(&rowId, &blockId, &sourceTxId, &sourceAddress, &destinationAddress, &asset, &amount, &status, &lastUpdatedBlockId, &txFee, &broadcastTxId, &paymentTag, &errorMessage); err == sql.ErrNoRows {
		payment = enulib.SimplePayment{}
		if err.Error() == "sql: no rows in result set" {
			payment.PaymentId = paymentId
			payment.Status = "Not found"
		}
	} else {
		payment = enulib.SimplePayment{SourceAddress: sourceAddress, DestinationAddress: destinationAddress, Asset: asset, Amount: amount, PaymentId: sourceTxId, Status: status, BroadcastTxId: broadcastTxId, TxFee: txFee, ErrorMessage: errorMessage}
	}

	return payment
}

func GetPaymentByPaymentTag(accessKey string, paymentTag string) enulib.SimplePayment {
	if isInit == false {
		Init()
	}

	//	 Query DB
	stmt, err := Db.Prepare("select rowId, blockId, sourceTxId, sourceAddress, destinationAddress, outAsset, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag, errorDescription from payments where paymentTag=? and accessKey=?")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(paymentTag, accessKey)
	if err != nil {
		panic(err.Error())
	}

	var rowId string
	var blockId string
	var sourceAddress string
	var destinationAddress string
	var asset string
	var amount uint64
	var txFee int64
	var broadcastTxId string
	var status string
	var sourceTxId string
	var lastUpdatedBlockId string
	var payment enulib.SimplePayment
	var errorMessage string

	if err := row.Scan(&rowId, &blockId, &sourceTxId, &sourceAddress, &destinationAddress, &asset, &amount, &status, &lastUpdatedBlockId, &txFee, &broadcastTxId, &errorMessage); err == sql.ErrNoRows {
		payment = enulib.SimplePayment{}
		if err.Error() == "sql: no rows in result set" {
			payment.PaymentTag = paymentTag
			payment.Status = "Not found"
		}
	} else {
		payment = enulib.SimplePayment{SourceAddress: sourceAddress, DestinationAddress: destinationAddress, Asset: asset, Amount: amount, PaymentId: sourceTxId, Status: status, BroadcastTxId: broadcastTxId, TxFee: txFee, ErrorMessage: errorMessage, PaymentTag: paymentTag}
	}

	return payment
}

func UpdatePaymentStatusByPaymentId(accessKey string, paymentId string, status string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set status=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(status, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdatePaymentWithErrorByPaymentId(accessKey string, paymentId string, errorDescription string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set status='error', errorDescription=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(errorDescription, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdatePaymentCompleteByPaymentId(accessKey string, paymentId string, txId string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set status='complete', broadcastTxId=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(txId, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

// create table userKeys (userId BIGINT, accessKey varchar(64), secret varchar(64), nonce bigint, assetId varchar(100), blockchainId varchar(100), sourceAddress varchar(100))
// Used to verify if the current request has a nonce > the value stored in the DB
func GetNonceByAccessKey(accessKey string) int64 {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select nonce from userkeys where accessKey=?")

	if err != nil {
		return -1
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var nonce int64
	row.Scan(&nonce)

	return nonce
}

// Used to retrieve the secret to verify the HMAC signature
func GetSecretByAccessKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select secret from userkeys where accessKey=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var secret string
	row.Scan(&secret)

	return secret
}

// Returns newest address associated with the access key
func GetSourceAddressByAccessKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select a.sourceAddress as sourceAddress from userKeys u left outer join addresses a on u.accessKey = a.accessKey where a.accessKey=? order by a.rowId desc limit 1;")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var sourceAddress string
	row.Scan(&sourceAddress)

	return sourceAddress
}

func GetAssetByAccessKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select assetId from userKeys where accessKey=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var assetId string
	row.Scan(&assetId)

	return assetId
}

// Used to update the value of the nonce after a successful API call
func UpdateNonce(accessKey string, nonce int64) error {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("update userkeys set nonce=? where accessKey=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(nonce, accessKey)
	if err2 != nil {
		return err2
	}

	return nil

}

func CreateUserKey(userId int64, assetId string, blockchainId string, sourceAddress string) (string, string, error) {
	if isInit == false {
		Init()
	}

	key := enulib.GenerateKey()
	secret := enulib.GenerateKey()

	// Open a transaction to ensure consistency between userKeys and addresses table
	tx, beginErr := Db.Begin()
	if beginErr != nil {
		return "", "", beginErr
	}

	// Insert into userKeys table first
	stmt, err := Db.Prepare("insert into userKeys(userId, accessKey, secret, nonce, assetId, blockchainId) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		tx.Rollback()
		return "", "", err
	}
	_, err = stmt.Exec(userId, key, secret, 0, assetId, blockchainId)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	// Insert into addresses second
	stmt2, err2 := Db.Prepare("insert into addresses(accessKey, sourceAddress) values(?, ?)")
	if err2 != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		tx.Rollback()
		return "", "", err
	}
	_, err2 = stmt2.Exec(key, sourceAddress)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return "", "", commitErr
	}

	defer stmt.Close()

	return key, secret, nil
}

func CreateSecondaryAddress(accessKey string, newAddress string) error {
	if isInit == false {
		Init()
	}

	// Check accessKey exists
	if UserKeyExists(accessKey) != true {
		log.Println("Call to CreateSecondaryAddress() with an invalid access key")

		return errors.New("Call to CreateSecondaryAddress() with an invalid access key")
	}

	stmt, err := Db.Prepare("insert into addresses(accessKey, sourceAddress) values(?, ?)")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		return err
	}

	// Perform the insert
	_, err = stmt.Exec(accessKey, newAddress)
	if err != nil {
		return err
	}

	defer stmt.Close()

	return nil
}

func UserKeyExists(accessKey string) bool {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select count(*) from userKeys where accessKey=?")

	if err != nil {
		return false
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var count int64
	row.Scan(&count)

	if count == 0 {
		return false
	}

	return true
}