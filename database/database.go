// database.go
package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/log"

	_ "github.com/vennd/enu/internal/github.com/go-sql-driver/mysql"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var Db *sql.DB
var databaseString string
var isInit bool = false // set to true only after the init sequence is complete

// Initialises global variables and database connection for all handlers
func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		log.Println("Found and using configuration file ./enuapi.json")
		configFilePath = "./enuapi.json"
	} else {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/bin/enuapi.json"
			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

		} else {
			if _, err := os.Stat(os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"); err == nil {
				configFilePath = os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"
				log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

			} else {
				log.Println("Cannot find enuapi.json")
				os.Exit(-100)
			}
		}
	}

	InitWithConfigPath(configFilePath)
}

func InitWithConfigPath(configFilePath string) {
	var configuration interface{}

	if isInit == true {
		return
	}

	// Read configuration from file
	log.Printf("Reading %s\n", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Println(err.Error())
		os.Exit(-100)
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

// Inserts an asset into the assets database
func InsertAsset(accessKey string, assetId string, sourceAddressValue string, assetValue string, descriptionValue string, quantityValue uint64, divisibleValue bool, status string) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into assets(accessKey, assetId, sourceAddress, asset, description, quantity, divisible, status) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(accessKey, assetId, sourceAddressValue, assetValue, descriptionValue, quantityValue, divisibleValue, status)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
}

func GetAssetByAssetId(c context.Context, accessKey string, assetId string) enulib.Asset {
	if isInit == false {
		Init()
	}

	// Set some initial values
	var assetStruct = enulib.Asset{}
	assetStruct.AssetId = assetId
	assetStruct.Status = "Not found"

	//	 Query DB
	log.FluentfContext(consts.LOGINFO, c, "select rowId, assetId, sourceAddress, asset, description, quantity, divisible, status, errorDescription, broadcastTxId from assets where assetId=%s and accessKey=%s", assetId, accessKey)
	stmt, err := Db.Prepare("select rowId, assetId, sourceAddress, asset, description, quantity, divisible, status, errorDescription,  broadcastTxId from assets where assetId=? and accessKey=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return assetStruct
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(assetId, accessKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to QueryRow. Reason: %s", err.Error())
		return assetStruct
	}

	var rowId string
	var sourceAddress string
	var asset string
	var description []byte
	var quantity uint64
	var divisible bool
	var status []byte
	var errorMessage []byte
	var broadcastTxId []byte

	if err := row.Scan(&rowId, &assetId, &sourceAddress, &asset, &description, &quantity, &divisible, &status, &errorMessage, &broadcastTxId); err == sql.ErrNoRows {
		if err.Error() == "sql: no rows in result set" {
		}
	} else if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
	} else {
		assetStruct = enulib.Asset{SourceAddress: sourceAddress, Asset: asset, Description: string(description), Quantity: quantity, AssetId: assetId, Status: string(status), ErrorMessage: string(errorMessage)}
	}

	return assetStruct
}

func UpdateAssetWithErrorByAssetId(c context.Context, accessKey string, assetId string, errorDescription string) error {
	if isInit == false {
		Init()
	}

	asset := GetAssetByAssetId(c, accessKey, assetId)

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update assets set status='error', errorDescription=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(errorDescription, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateAssetStatusByAssetId(c context.Context, accessKey string, assetId string, status string) error {
	if isInit == false {
		Init()
	}

	asset := GetAssetByAssetId(c, accessKey, assetId)

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update assets set status=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(status, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateAssetNameByAssetId(c context.Context, accessKey string, assetId string, assetName string) error {
	if isInit == false {
		Init()
	}

	asset := GetAssetByAssetId(c, accessKey, assetId)

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update assets set asset=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(assetName, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateAssetCompleteByAssetId(c context.Context, accessKey string, assetId string, txId string) error {
	if isInit == false {
		Init()
	}

	asset := GetAssetByAssetId(c, accessKey, assetId)

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	log.Printf("update assets set status='complete', broadcastTxId=%s where accessKey=%s and assetId = %s\n", txId, accessKey, assetId)

	stmt, err := Db.Prepare("update assets set status='complete', broadcastTxId=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(txId, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

// Inserts a dividend into the dividends database
func InsertDividend(accessKey string, dividendId string, sourceAddressValue string, assetValue string, dividendAssetValue string, quantityPerUnitValue uint64, status string) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into dividends(accessKey, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, status) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(accessKey, dividendId, sourceAddressValue, assetValue, dividendAssetValue, quantityPerUnitValue, status)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
}

func GetDividendByDividendId(c context.Context, accessKey string, dividendId string) enulib.Dividend {
	if isInit == false {
		Init()
	}

	// Initialise some initial values
	var dividendStruct = enulib.Dividend{}
	dividendStruct.DividendId = dividendId
	dividendStruct.Status = "Not found"

	//	 Query DB
	log.FluentfContext(consts.LOGERROR, c, "select rowId, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, errorDescription, broadcastTxId from dividends where dividendId=%s and accessKey=%s", dividendId, accessKey)
	stmt, err := Db.Prepare("select rowId, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, status, errorDescription, broadcastTxId from dividends where dividendId=? and accessKey=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return dividendStruct
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(dividendId, accessKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to QueryRow. Reason: %s", err.Error())
		return dividendStruct
	}

	var rowId string
	var sourceAddress string
	var asset string
	var dividendAsset string
	var quantityPerUnit uint64
	var status string
	var errorMessage string
	var broadcastTxId []byte

	if err := row.Scan(&rowId, &dividendId, &sourceAddress, &asset, &dividendAsset, &quantityPerUnit, &status, &errorMessage, &broadcastTxId); err == sql.ErrNoRows {
		if err.Error() == "sql: no rows in result set" {
			return dividendStruct
		}
	} else if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
	} else {
		dividendStruct = enulib.Dividend{SourceAddress: sourceAddress, Asset: asset, DividendAsset: dividendAsset, QuantityPerUnit: quantityPerUnit, DividendId: dividendId, Status: status, ErrorMessage: errorMessage, BroadcastTxId: string(broadcastTxId)}
	}

	return dividendStruct
}

func UpdateDividendWithErrorByDividendId(c context.Context, accessKey string, dividendId string, errorDescription string) error {
	if isInit == false {
		Init()
	}

	dividend := GetDividendByDividendId(c, accessKey, dividendId)

	if dividend.DividendId == "" {
		errorString := fmt.Sprintf("Dividend does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update dividends set status='error', errorDescription=? where accessKey=? and dividendId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(errorDescription, accessKey, dividendId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateDividendCompleteByDividendId(c context.Context, accessKey string, dividendId string, txId string) error {
	if isInit == false {
		Init()
	}

	dividend := GetDividendByDividendId(c, accessKey, dividendId)

	if dividend.DividendId == "" {
		errorString := fmt.Sprintf("Dividend does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	log.FluentfContext(consts.LOGINFO, c, "update dividends set status='complete', broadcastTxId=%s where accessKey=%s and dividendId = %s\n", txId, accessKey, dividendId)

	stmt, err := Db.Prepare("update dividends set status='complete', broadcastTxId=? where accessKey=? and dividendId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(txId, accessKey, dividendId)
	if err2 != nil {
		return err2
	}

	return nil
}

// Inserts a payment into the payment database
func InsertPayment(accessKey string, blockIdValue int64, sourceTxidValue string, sourceAddressValue string, destinationAddressValue string, outAssetValue string, outAmountValue uint64, statusValue string, lastUpdatedBlockIdValue int64, txFeeValue uint64, paymentTag string, requestId string) {
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

func UpdatePaymentSignedRawTxByPaymentId(accessKey string, paymentId string, signedRawTx string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set signedRawTx=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(signedRawTx, accessKey, paymentId)
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

	stmt, err := Db.Prepare("select assetId from userKeys where accessKey=? and status=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey, consts.AccessKeyValidStatus)

	var assetId string
	row.Scan(&assetId)

	return assetId
}

// Used to update the value of the nonce after a successful API call
func UpdateNonce(accessKey string, nonce int64) error {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("update userkeys set nonce=? where accessKey=? and status=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(nonce, accessKey, consts.AccessKeyValidStatus)
	if err2 != nil {
		return err2
	}

	return nil

}

func CreateUserKey(userId int64, assetId string, blockchainId string, sourceAddress string, parentAccessKey string) (string, string, error) {
	if isInit == false {
		Init()
	}

	// blockchainId must be in the list of blockchains that we support
	supportedBlockchains := consts.SupportedBlockchains
	sort.Strings(supportedBlockchains)

	i := sort.SearchStrings(supportedBlockchains, blockchainId)
	blockchainValid := i < len(supportedBlockchains) && supportedBlockchains[i] == blockchainId

	if blockchainValid == false {
		e := fmt.Sprintf("Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))

		return "", "", errors.New(e)
	}

	key := enulib.GenerateKey()
	secret := enulib.GenerateKey()

	// Open a transaction to ensure consistency between userKeys and addresses table
	tx, beginErr := Db.Begin()
	if beginErr != nil {
		return "", "", beginErr
	}

	// Insert into userKeys table first
	stmt, err := Db.Prepare("insert into userkeys(userId, parentAccessKey, accessKey, secret, nonce, assetId, blockchainId, status) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		tx.Rollback()
		return "", "", err
	}
	_, err = stmt.Exec(userId, parentAccessKey, key, secret, 0, assetId, blockchainId, consts.AccessKeyValidStatus)
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

func CreateSecondaryAddress(accessKey string, newAddress string, requestId string) error {
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

// Only return true where an accessKey exists and also has a valid status
func UserKeyExists(accessKey string) bool {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select count(*) from userkeys where accesskey=? and status=?")

	if err != nil {
		return false
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey, consts.AccessKeyValidStatus)

	var count int64
	row.Scan(&count)

	if count == 0 {
		return false
	}

	return true
}

// Updates a given accessKey, ignores what the existing status is
func UpdateUserKeyStatus(accessKey string, status string) error {
	if isInit == false {
		Init()
	}

	// status must be a supported access key status
	statuses := consts.AccessKeyStatuses
	sort.Strings(statuses)

	x := sort.SearchStrings(statuses, status)
	statusValid := x < len(statuses) && statuses[x] == status
	if statusValid == false {
		e := fmt.Sprintf("Attempt to update status to an invalid value: %s. Valid values: %s", status, strings.Join(statuses, ", "))

		return errors.New(e)
	}

	stmt, err := Db.Prepare("update userkeys set status=? where accessKey=?")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		return err
	}

	// Perform the update
	_, err = stmt.Exec(status, accessKey)
	if err != nil {
		return err
	}

	defer stmt.Close()

	return nil
}

func GetStatusByUserKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select status from userkeys where accesskey=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var status string
	row.Scan(&status)

	return status
}

func GetBlockchainIdByUserKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select blockchainId from userkeys where accesskey=? and status=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey, consts.AccessKeyValidStatus)

	var blockchainId string
	row.Scan(&blockchainId)

	return blockchainId
}
