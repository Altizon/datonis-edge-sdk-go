package edge

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	//"regexp"
	"bytes"
	"strings"
	"sync"
	"time"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	QoS0 = iota
	QoS1
	QoS2
)

var MissingKey = errors.New("Cannot subscribe to instructions as thing or device key not present.")
var MqttClientId string;

type Instruction struct {
	ThingKey           string `json:"thing_key,omitempty"`
	DeviceKey          string `json:"device_key,omitempty"`
	AlertKey           string `json:"alert_key,omitempty"`
	InstructionWrapper struct {
		Instruction map[string]interface{} `json:"instruction,omitempty"`
	} `json:"instruction_wrapper,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	Hash      string `json:"hash,omitempty"`
	Topic     string `json:"-"`
}

type MqttGateway struct {
	*GatewayConfig

	Queue                sync.WaitGroup // to wait on messages.
	Acks                 chan string
	Instructions         chan Instruction
	InstructionHandler	 func (gw EdgeGateway, ts int64, t *Thing, alertKey string, instruction map[string]interface{})
	client               MQTT.Client
	RegisteredKeys       []string
	RegisteredDeviceKeys []string
	SubscribedTopics     []string
}

func (gw *MqttGateway) Wait() {
	gw.Queue.Wait()
}

func NewClient(config *GatewayConfig) MQTT.Client {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Url())
	MqttClientId = config.AccessKey[0:10] + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	opts.SetClientID(MqttClientId)

	return MQTT.NewClient(opts)
}

func ConnectMqtt(config *GatewayConfig) (*MqttGateway, error) {
	client := NewClient(config)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Printf("Gateway connection: %v\n", client.IsConnected())
	gw := MqttGateway{
		GatewayConfig:        config,
		Acks:                 make(chan string),
		Instructions:         make(chan Instruction),
		client:               client,
		SubscribedTopics:     make([]string, 0, 25),
		RegisteredKeys:       make([]string, 0, 25),
		RegisteredDeviceKeys: make([]string, 0, 25),
	}

	go instructionWorker("InstructionWorker-Thread", &gw)

	var callback MQTT.MessageHandler = func(c MQTT.Client, msg MQTT.Message) {
		// Ack it
		gw.Acks <- string(msg.Payload())
	}

	http_ack := fmt.Sprintf("Altizon/Datonis/%s/httpAck", MqttClientId)
	token = client.Subscribe(http_ack, 1, callback)
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return &gw, nil
}

func (gw *MqttGateway) IsConnected() bool {
	return gw.client.IsConnected()
}

func (gw *MqttGateway) GetConfig() *GatewayConfig {
	return gw.GatewayConfig
}

func instructionWorker(threadName string, gw *MqttGateway) {
	fmt.Println("Started: ", threadName)
	for true {
		ins := <-gw.Instructions
		if gw.InstructionHandler != nil {
			fmt.Println("Recieved new Instruction: ", ins)
			t := NewThing()
			t.Key = ins.ThingKey
			t.DeviceKey = ins.DeviceKey
			gw.InstructionHandler(gw, ins.Timestamp, t, ins.AlertKey, ins.InstructionWrapper.Instruction)
		} else {
			fmt.Println("Received instruction: ", ins, ". But, no handler set for executing it. It will be ignored")
		}
	}
	fmt.Println("Stopped: ", threadName)
}

func (gw *MqttGateway) SubscribeForInstructions(t *Thing) error {
	var key string
	if t.Key != "" {
		key = t.Key
	} else if t.DeviceKey != "" {
		key = t.DeviceKey
	} else {
		return MissingKey
	}

	topic := fmt.Sprintf("Altizon/Datonis/%s/thing/%s/executeInstruction", gw.AccessKey, key)

	fmt.Println("Sending -", topic)
	token := gw.client.Subscribe(topic, QoS2, func(c MQTT.Client, msg MQTT.Message) {
		var idx int

		rawPayload := msg.Payload()
		payload := string(rawPayload)

		// get hash from payload
		hash := regexp.MustCompile("(.?){64}\"}$").FindString(payload)
		if len(hash) > 2 {
			hash = hash[:64]
		}

		// get data to compute HMAC
		idx = strings.Index(payload, ",\"access_key")
		var data = make([]byte, idx)
		copy(data, rawPayload[:idx])
		data = append(data, '}')

		sig := encode(gw.SecretKey, data)
		// reject instruction if HMAC doesn't match
		if sig == hash {
			ins := Instruction{Topic: topic}
			buf := bytes.NewReader(rawPayload)
			dec := json.NewDecoder(buf)
			dec.Decode(&ins)
			gw.Instructions <- ins
		} else {
			fmt.Println("Hash code for the instruction from the server does not match with the re-calculated one. Ignoring instruction!")
		}
	})

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	gw.SubscribedTopics = append(gw.SubscribedTopics, topic)
	fmt.Println("Subscribed to ", topic)
	return nil
}

func (gw *MqttGateway) ThingRegister(t *Thing) error {
	if stringInSlice(gw.RegisteredKeys, t.Key) || stringInSlice(gw.RegisteredDeviceKeys, t.DeviceKey) {
		return nil
	}
	payload := CreateThingRegister(t, 0)
	topic := fmt.Sprintf("Altizon/Datonis/%s/register", MqttClientId)
	if err := gw.sync_send_message(topic, payload, 1); err != nil {
		return err
	} else {
		gw.RegisteredKeys = append(gw.RegisteredKeys, t.Key)
		gw.RegisteredDeviceKeys = append(gw.RegisteredDeviceKeys, t.DeviceKey)
		return gw.SubscribeForInstructions(t)
	}
}

func (gw *MqttGateway) ThingHeartbeat(t *Thing, ts int64) error {
	eventData := CreateThingHeartbeat(t, ts)
	topic := fmt.Sprintf("Altizon/Datonis/%s/heartbeat", MqttClientId)
	return gw.sync_send_message(topic, eventData, 0)
}

func (gw *MqttGateway) ThingEvent(eventData map[string]interface{}) error {
	topic := fmt.Sprintf("Altizon/Datonis/%s/event", MqttClientId)
	return gw.sync_send_message(topic, eventData, 1)
}

func (gw *MqttGateway) BulkThingEvent(eventData []map[string]interface{}) error {
	bulkEventData := map[string]interface{} {
		"events": eventData,
	}
	return gw.ThingEvent(bulkEventData)
}

func (gw *MqttGateway) SetInstructionHandle(instructionHandler func(gw EdgeGateway, ts int64, t *Thing, alertKey string, instruction map[string]interface{})) error {
	gw.InstructionHandler = instructionHandler
	return nil
}

func (gw *MqttGateway) InstructionAck(alertKey, message string, alertLevel int, data map[string]interface{}) error {
	topic := fmt.Sprintf("Altizon/Datonis/%s/alert", MqttClientId)
	payload := CreateInstructionAlert(alertKey, message, alertLevel, data, 0)
	return gw.sync_send_message(topic, payload, 1)
}

func (gw *MqttGateway) Alert(t *Thing, alertMessage string, alertLevel int, alertData map[string]interface{}) error {
	topic := fmt.Sprintf("Altizon/Datonis/%s/alert", MqttClientId)
	payload := CreateAlert(t, alertMessage, alertLevel, alertData, 0)
	return gw.sync_send_message(topic, payload, 1)
}

func (gw *MqttGateway) sync_send_message(topic string, payload map[string]interface{}, qos byte) error {
	fmt.Printf("Sending - %s\n", topic)

	hash, _ := json.Marshal(payload)
	var buffer bytes.Buffer
	buffer.WriteString(strings.TrimSuffix(string(hash), "}"))
	buffer.WriteString(",\"hash\":\"")
	sig := string(encode(gw.SecretKey, hash))
	buffer.WriteString(sig)
	buffer.WriteString("\"")
	buffer.WriteString(",\"access_key\":\"")
	buffer.WriteString(string(gw.AccessKey))
	buffer.WriteString("\"}")

	message := []byte(buffer.String())
	fmt.Println(string(message)) // end: Debug and Print
	context := sig

	if token := gw.client.Publish(topic, qos, false, message); token.Wait() && token.Error() != nil {
		// No point in panic here. Log this message and move on!
		fmt.Println("Error: ", token.Error())
	}

	/* Loop over the channel waiting for data */
	for value := range gw.Acks {
		var m map[string]interface{}
		err := json.Unmarshal([]byte(value), &m)
		if (err == nil) && (m["context"] == context) {
			//fmt.Printf("Context ----> %s\n", m["context"])
			// My message was ack'ed.
			// Check for return code & then I'm outta here
			var c, _ = m["http_code"].(float64)
			code := int(c)
			//fmt.Printf("\nHTTP Code ----> %d, OK --> %s\n", code, ok)
			if code != 200 {
				return errors.New(fmt.Sprintf("There was an error with the request. Response code: %d", code))
			}
			return nil
		} else {
			// Ignore.put the data back in the channel
			fmt.Println("ignore: ", value)
			gw.Acks <- value
		}
	}
	return nil
}

func (gw *MqttGateway) send_message(topic string, data map[string]interface{}, qos byte) {
	gw.Queue.Add(1)

	go func() {
		defer gw.Queue.Done()
		if err := gw.sync_send_message(topic, data, qos); err != nil {
			fmt.Printf("Error: %s. What can we do here?\n", err)
		}
	}()
}
