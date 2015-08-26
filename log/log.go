// Forwards all logging to Fluentd
package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/vennd/enu/consts"

	"github.com/vennd/enu/internal/github.com/nytlabs/gojsonexplode"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

// Globals
var isInit bool = false // set to true only after the init sequence is complete
var fluentHost string

type logObject struct {
	Tag         string `json:"tag"`
	ErrorString string `json:"errorString"`
	Object      interface{}
}

// Initialises global variables and database connection for all handlers
func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		//		log.Println("Found and using configuration file ./enuapi.json")
		configFilePath = "./enuapi.json"
	} else {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/bin/enuapi.json"
			//			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

		} else {
			if _, err := os.Stat(os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"); err == nil {
				configFilePath = os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"
				//				log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)
			} else {
				log.Fatalln("Cannot find enuapi.json")
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
	//	log.Printf("Reading %s\n", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Fatalln(err)
	}

	err = json.Unmarshal(file, &configuration)

	if err != nil {
		log.Println("Unable to parse enuapi.json")
		log.Fatalln(err)
	}

	m := configuration.(map[string]interface{})

	// Fluentd http forwarder parameters
	fluentHost = m["fluentHost"].(string)

	isInit = true
}

// Compatibility function with existing logger.
// Writes a copy of the string to format to stdout but also sends a copy to Fluent
// Uses a default tag of 'enu.$ENV.$HOSTNAME'
// Note: If unable to forward to Fluent, this function will NOT raise errors with respect to Fluent
func Printf(format string, a ...interface{}) {
	fluentf("", true, format, a...)
}

// Compatibility function with existing logger.
// Writes a copy of the string to format to stdout but also sends a copy to Fluent
// Uses a default tag of 'enu.$ENV.$HOSTNAME'
// Note: If unable to forward to Fluent, this function will NOT raise errors with respect to Fluent
func Println(a string) {
	fluentf("", true, "%s", a)
}

// Log a formatted string to Fluent.
// It is suggested that 'tag' be set to the name of the source file. eg "log.go"
// Otherwise, 'tag' can be set to an empty string if the default tag of 'enu.$ENV.$HOSTNAME' is sufficient
// Use this function whenever doing general logging which doesn't require the context to be logged
func Fluentf(tag string, format string, a ...interface{}) {
	fluentf(tag, false, format, a...)
}

// Log string to Fluent.
// It is suggested that 'tag' be set to the name of the source file. eg "log.go"
// Otherwise, 'tag' can be set to an empty string if the default tag of 'enu.$ENV.$HOSTNAME' is sufficient
// Use this function whenever doing general logging which doesn't require the context to be logged
func Fluentln(tag string, a string) {
	fluentf(tag, false, "%s", a)
}

func fluentf(tag string, compatibilityMode bool, format string, a ...interface{}) {
	errorString := fmt.Sprintf(format, a...)

	env := os.Getenv("ENV")
	hostname, err := os.Hostname()

	if err != nil {
		hostname = "unknown"
	}

	if env == "" {
		env = "unknown"
	}

	if compatibilityMode {
		log.Printf(format, a...)
	}

	fullTag := "enu." + env + "." + hostname
	if tag != "" {
		fullTag += "." + tag
	} else {
	}

	object(fullTag, nil, errorString, compatibilityMode)
}

// Log a formatted string with a corresponding context to Fluent.
// The values from the context are copied to a local struct
func FluentfContext(tag string, context context.Context, format string, a ...interface{}) {
	type contextValues struct {
		RequestId    string `json:"requestId"`
		BlockchainId string `json:"blockchainId"`
		AccessId     string `json:"accessId"`
		Nonce        int64  `json:"nonce"`
	}

	var objectToLog contextValues
	objectToLog.RequestId = context.Value(consts.RequestIdKey).(string)
	objectToLog.BlockchainId = context.Value(consts.BlockchainIdKey).(string)
	objectToLog.AccessId = context.Value(consts.AccessKeyKey).(string)
	objectToLog.Nonce = context.Value(consts.NonceIntKey).(int64)

	errorString := fmt.Sprintf(format, a...)

	env := os.Getenv("ENV")
	hostname, err := os.Hostname()

	if err != nil {
		hostname = "unknown"
	}

	if env == "" {
		env = "unknown"
	}

	fullTag := "enu." + env + "." + hostname
	if tag != "" {
		fullTag += "." + tag
	}

	object(fullTag, objectToLog, errorString, false)
}

// Serialises the given object into JSON and then sends to Fluent via the HTTP forwarder
func object(tag string, object interface{}, errorString string, suppressErrors bool) {
	var LogObject logObject
	var payloadJsonBytes []byte
	var err error

	LogObject.ErrorString = errorString
	LogObject.Tag = tag
	LogObject.Object = object

	if isInit == false {
		Init()
	}

	payloadJsonBytes, err = json.Marshal(LogObject)

	if err != nil {
		logString := fmt.Sprintf("log.go: Fatal error - unable to marshall to json: %s", object)
		log.Println(logString)
	}

	_, err2 := sendToFluent(fluentHost+"/"+tag, payloadJsonBytes)

	// If running in suppressErrors mode, don't raise if we couldn't send to fluentd
	// suppressErrors mode is used when Object is called by Printf for backwards compatibility
	if err2 != nil && suppressErrors == false {
		log.Println("log.go: Fatal error - failed to send to fluentd")
		log.Printf("\"%s\",\"%s\",\"%s\"\n", errorString, tag, string(payloadJsonBytes)) // fallback to printing to stdout
	}
}

func sendToFluent(url string, postData []byte) (int64, error) {
	var flattenedPostData []byte
	var err error
	var postDataJson string

	if len(postData) != 0 {
		flattenedPostData, err = gojsonexplode.Explodejson(postData, ".")
	} else {
		flattenedPostData = make([]byte, 0)
	}

	if err != nil {
		logString := fmt.Sprintf("log.go: Fatal error - unable to flatten json: %s", string(flattenedPostData))
		log.Println(logString)
		return -1, err
	}

	if flattenedPostData != nil {
		postDataJson = string(flattenedPostData)
	} else {
		postDataJson = ""
	}

	// Set headers
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(postDataJson))
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)

	if err != nil {
		return -1, err
	}

	return int64(resp.StatusCode), nil
}