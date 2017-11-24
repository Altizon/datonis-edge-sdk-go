package edge

type EdgeGateway interface {

	Wait()

	GetConfig() *GatewayConfig

	// Registers the specified thing with Datonis
	ThingRegister(t *Thing) error

	// Sends a Heart Beat message to Datonis indicating that this thing is alive
	ThingHeartbeat(t *Thing, ts int64) error

	// Sends a Thing Data Packet (event) to Datonis
	ThingEvent(eventData map[string]interface{}) error
	// Sends the multiple Thing Data Packet (BulkEvents) to Datonis
	BulkThingEvent(eventData []map[string]interface{}) error

	// Sends the Alerts to Datonis
	Alert(t *Thing, message string, alertLevel int, data map[string]interface{}) error

	// Set the instruction handler to handle the instruciton received
	SetInstructionHandle(instructionHandler func(gw EdgeGateway, ts int64, t *Thing, alertKey string, instruction map[string]interface{})) error

	// Sends the instruction acknowledgement to datonis through MQTT protocol.
	InstructionAck(alertKey, message string, alertLevel int, data map[string]interface{}) error
}

func CreateGateway(c *GatewayConfig) EdgeGateway {

	if c.Protocol == "mqtt" || c.Protocol == "mqtts" {
		gw, _ := ConnectMqtt(c)
		return gw
	} else {
		gw, _ := ConnectHttp(c)
		return gw
	}
}

