package edge

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	// "regexp"
	"bytes"
	// "strings"
	"sync"
	"net/http"
	"io/ioutil"
	// "time"
	// "strconv"
)

type HttpGateway struct {
	*GatewayConfig

	Queue                sync.WaitGroup // to wait on messages.
}

func (gw *HttpGateway) Wait() {
	gw.Queue.Wait()
}

func ConnectHttp(config *GatewayConfig) (*HttpGateway, error) {
	gw := HttpGateway{
		GatewayConfig:        config,
	}
	return &gw, nil
}

func (gw *HttpGateway) GetConfig() *GatewayConfig {
	return gw.GatewayConfig
}

func (gw *HttpGateway) ThingRegister(t *Thing) error {
	return errors.New("Thing registration through HTTPGateway not required.")
}

func (gw *HttpGateway) ThingHeartbeat(t *Thing, ts int64) error {
	eventData := CreateThingHeartbeat(t, ts)
	retVal := postMessage(gw.GatewayConfig, "/api/v3/things/heartbeat.json", eventData)
	return retVal
}

func (gw *HttpGateway) ThingEvent(eventData map[string]interface{}) error {
	retVal := postMessage(gw.GatewayConfig, "/api/v3/things/event.json", eventData)
	return retVal
}

func (gw *HttpGateway) BulkThingEvent(eventData []map[string]interface{}) error {
	bulkEventData := map[string]interface{} {
		"events": eventData,
	}
	return gw.ThingEvent(bulkEventData)
}

func (gw *HttpGateway) Alert(t *Thing, alertMessage string, alertLevel int, alertData map[string]interface{}) error {
	payload := CreateAlert(t, alertMessage, alertLevel, alertData, 0)
	retVal := postMessage(gw.GatewayConfig, "/api/v3/alerts.json", payload)
	return retVal
}

func (gw *HttpGateway) SetInstructionHandle(instructionHandler func(gw EdgeGateway, ts int64, t *Thing, alertKey string, instruction map[string]interface{})) error {
	return errors.New("Instruction handler for HTTPGateway not required.")
}

func (gw *HttpGateway) InstructionAck(alertKey, message string, alertLevel int, data map[string]interface{}) error {
	return errors.New("Instruction acknowledgement through HTTPGateway not required.")
}

func postMessage(c *GatewayConfig, url string, payload map[string]interface{}) error {
	
	hash, _ := json.Marshal(payload)
	sig := string(encode(c.SecretKey, hash))
	postUrl := c.Url() + url
	req, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(hash))
	req.Header.Set("X-Dtn-Signature", sig)
	req.Header.Set("X-Access-Key", string(c.AccessKey))
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("Sending - ", postUrl)
	fmt.Println(payload)
	client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 && string(body) != "" {
		var m interface{}
		err := json.Unmarshal([]byte(body), &m)
		if err == nil {
			var errs []interface{}
			if reflect.ValueOf(m).Kind() == reflect.Map {
				errs = m.(map[string]interface{})["errors"].([]interface{})
			} else {
				errs = m.([]interface{})
			}
			for _, errMsg := range errs {
				parsedErrMsg := errMsg.(map[string]interface{})
				fmt.Println("Error: " + parsedErrMsg["code"].(string) + " : " + parsedErrMsg["message"].(string))
			}
		} else {
			fmt.Println("Error in parsing the json body: ", string(body))
		}
	} else if resp.StatusCode == 200 {
		return nil
	}
	fmt.Println("Response Body:", string(body))

	return errors.New(fmt.Sprintf("There was an error with the request. Response code: %d", resp.StatusCode))
}