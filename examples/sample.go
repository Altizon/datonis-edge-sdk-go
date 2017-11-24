package main

import (
	"fmt"
	"math/rand"
	"time"
	"../edge"
)

const BULK_MAX_ELEMENT = 5
// Set BULK_TRANSMIT = true if want to send data in bulk.
const BULK_TRANSMIT = false
var MAXEVENTS = 100
var EVENT_FREQUENCY = 360 // events/hour
var bulkEvents = []map[string]interface{} {}

func addEvent(eventData map[string]interface{}) {
	bulkEvents = append(bulkEvents, eventData)
}

func clearEvents() {
	bulkEvents = bulkEvents[0:0]
}

func nQueuedEvents() int {
	return len(bulkEvents)
}

func getData() map[string]interface{} {
	return map[string]interface{} {
		"temperature": rand.Intn(60),
		"pressure": rand.Intn(30),
	}
}

func getWaypoint() []float64 {
	// waypoint format: [latitude, longitude]
	return []float64{(18.0 + rand.Float64()), (73.0 + rand.Float64())}
}

func sendEvent(gw edge.EdgeGateway, t *edge.Thing) {
	// Send Event data
	var data map[string]interface{}
	// Do not initialize the data if only want to send waypoints
	data = getData()

	var waypoint []float64
	// Uncomment below line to send waypoint
	// waypoint = getWaypoint()

	eventData := edge.CreateThingEvent(t, data, waypoint, 0)
	if BULK_TRANSMIT {
		addEvent(eventData)
		if nQueuedEvents() >= BULK_MAX_ELEMENT  {
			retVal := gw.BulkThingEvent(bulkEvents)
			if retVal == nil {
				fmt.Println("Bulk send thing events: ", bulkEvents)
			} else {
				fmt.Println("Unable to bulk send thing events: ", retVal)
			}
			clearEvents()
		} else {
			fmt.Println("Thing Event Queued: ", eventData)
		}
	} else {
		retVal := gw.ThingEvent(eventData)
		if retVal == nil {
			fmt.Println("Send thing event: ", eventData)
		} else {
			fmt.Println("Unable to send thing event: ", retVal)
		}
	}
}

func sendAlert(gw edge.EdgeGateway, t *edge.Thing, alertLevel int, alertLevelStr string) error {
	data := map[string]interface{}{"tempearture": 120}
	if err := gw.Alert(t, "This is an example " + alertLevelStr + " alert using go SDK.", alertLevel, data); err != nil {
		fmt.Println("Unable to send Alert: ", err)
	}
	return nil
}

func sendAlerts(gw edge.EdgeGateway, t *edge.Thing) error {
	sendAlert(gw, t, 0, "INFO")
    sendAlert(gw, t, 1, "WARNING")
    sendAlert(gw, t, 2, "ERROR")
	sendAlert(gw, t, 3, "CRITICAL")
	return nil
}

func handleInstruciton(gw edge.EdgeGateway, ts int64, t *edge.Thing, alertKey string, instruction map[string]interface{}) {
	fmt.Println("Recieved Instruction Data for thing with key-device_key ", t.Key, "-", t.DeviceKey, ": ", instruction)
	alertMsg := fmt.Sprintf("Instruction with alert key %s executed", alertKey)
	if err := gw.InstructionAck(alertKey, alertMsg, 0, map[string]interface{}{"execution_status":"success"}); err == nil {
		fmt.Println("Sent instruction execution ACK back to datonis")
	} else {
		fmt.Println("Could not send instruction execution ACK back to datonis")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Here we go..")
	config := &edge.GatewayConfig{
		AccessKey: "1d2fb5c369863fd54afafde654c26dtd51122t8e",
		SecretKey: "f4e31122629etaeaa48d9c8c72b8cctfc9d63acc",
	}
	// Change the Datonis host and port using below config change.
	// config.Host = "localhost"
	// config.Port = 3000

	// Choose the protocol to send the data.
	// config.Protocol = "mqtts"
	// config.Protocol = "mqtt"
	config.Protocol = "http"
	// config.Protocol = "https"

	config = edge.InitGatewayConfig(config)
	gw := edge.CreateGateway(config)

	t := edge.NewThing()
	t.Name = "Compressor"
	t.Key = "614a5ed34c"

	// Register thing if using the mqtt or mqtts protocol.
	// t.BiDirectional = true
	// if err := gw.ThingRegister(t); err != nil {
	// 	fmt.Println("Unable to Register: ", err)
	// }

	// Handle instructions if using the mqtt or mqtts protocol.
	// if err := gw.SetInstructionHandle(handleInstruciton); err == nil {
	// 	fmt.Println("Started handling the instruction.")
	// } else {
	// 	fmt.Println("Could not started handling the instruction: ", err)		
	// }

	// Send Alerts
	sendAlerts(gw, t)

	count := 0
	for i := 0; i < MAXEVENTS; i++ {
		if count == 0 {
			// Send a heartbeat.
			gw.ThingHeartbeat(t, 0)
		}
		sendEvent(gw, t)
		time.Sleep(time.Duration(3600000/EVENT_FREQUENCY) * time.Millisecond)
		count++
		count = count % 10
	}

	fmt.Println("Waiting...")
	gw.Wait()
	fmt.Println("Bye Bye")
}
