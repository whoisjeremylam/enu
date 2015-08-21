// Forwards all logging to Fluentd
package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/nytlabs/gojsonexplode"
	"golang.org/x/net/context"
)

// Globals
var isInit bool = false // set to true only after the init sequence is complete
var fluentHost string

type logObject struct {
	Tag string `json:"tag"`
	ErrorString  string `json:"errorString"`
	Object interface{}
}

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
	log.Printf("Reading %s\n", configFilePath)
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

// Compatibility function with existing logger. Sends to fluent instead using a default tag of 'enu'
func Printf(format string, a ...interface{}) {
	errorString := fmt.Sprintf(format, a)
	
	Object("enu", nil, errorString)
}

// Serialises the given object into JSON and then sends to Fluent via the HTTP forwarder
func Object(tag string, object interface{}, errorString string) {
	var LogObject logObject
	var payloadJsonBytes []byte
	var err error
	
	LogObject.ErrorString = errorString
	LogObject.Tag = tag
	LogObject.Object = object
	
	if isInit == false {
		Init()
	}

	if object == nil {
		payloadJsonBytes = make([]byte, 0)
	} else {
		payloadJsonBytes, err = json.Marshal(LogObject)
	}
	

	if err != nil {
		logString := fmt.Sprintf("log.go: Unable to marshall to json: %s", object)
		log.Println(logString)
	}

	_, err2 := sendToFluent(fluentHost+"/"+tag, payloadJsonBytes)

	if err2 != nil {
		logString := fmt.Sprintf("log.go: Unable to send to fluentd: %s", err2.Error())
		log.Println(logString)
	}
}

// Serialises the given object into JSON and then sends to Fluent via the HTTP forwarder
func ContextAndObject(tag string, context context.Context, object interface{}) {

}

func sendToFluent(url string, postData []byte) (int64, error) {
	flattenedPostData, err := gojsonexplode.Explodejson(postData, ".")

	if err != nil {
		logString := fmt.Sprintf("log.go: Unable to flatten json: %s", string(flattenedPostData))
		log.Println(logString)
		return -1, errors.New(logString)
	}

	postDataJson := string(flattenedPostData)

	// Set headers
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(postDataJson))
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)

	return int64(resp.StatusCode), nil
}
