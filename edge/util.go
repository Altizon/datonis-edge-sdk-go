package edge

import (
	"crypto/hmac"
	"crypto/sha256"
	//"encoding/base64"
	"encoding/hex"
	"time"
)

func GetCurrentTS() int64 {
	return (time.Now().UnixNano() / 1000000)
}

func encode(key string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(payload)
	//return base64.StdEncoding.EncodeToString(h.Sum(nil))
	return hex.EncodeToString(h.Sum(nil))
}

func stringInSlice(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func CreateThingHeartbeat(t *Thing, ts int64) map[string]interface{} {
	if ts == 0 {
		ts = GetCurrentTS()
	}
	data := map[string]interface{}{
		"timestamp": ts,
		"thing_key": t.Key,
		"device_key": t.DeviceKey,
	}

	return data
}

func CreateThingRegister(t *Thing, ts int64) map[string]interface{} {
	if ts == 0 {
		ts = GetCurrentTS()
	}
	data := map[string]interface{}{
		"timestamp":		ts,
		"thing_key":		t.Key,
		"device_key":		t.DeviceKey,
		"bi_directional": 	t.BiDirectional,
	}
	if t.Name != "" {
		data["name"] = t.Name
	}
	if t.Description != "" {
		data["description"] = t.Description
	}

	return data
}

func CreateThingEvent(t *Thing, data map[string]interface{}, waypoint []float64, ts int64) map[string]interface{} {
	if ts == 0 {
		ts = GetCurrentTS()
	}
	eventData := map[string]interface{} {
		"timestamp": ts,
	}
	if data != nil {
		eventData["data"] = data
	}
	if waypoint != nil {
		eventData["waypoint"] = waypoint
	}
	if t.Key != "" {
		eventData["thing_key"] = t.Key
	}
	if t.DeviceKey != "" {
		eventData["device_key"] = t.DeviceKey
	}

	return eventData
}

func CreateAlert(t *Thing, alertMessage string, alertLevel int, alertData map[string]interface{}, ts int64) map[string]interface{} {
	if ts == 0 {
		ts = GetCurrentTS()
	}
	payload := map[string]interface{}{
		"alert": map[string]interface{}{
			"data":       alertData,
			"message":    alertMessage,
			"alert_type": alertLevel,
			"thing_key":  t.Key,
			"device_key": t.DeviceKey,
			"timestamp":  ts,
		},
	}

	return payload
}

func CreateInstructionAlert(alertKey string, alertMessage string, alertLevel int, alertData map[string]interface{}, ts int64) map[string]interface{} {
	if ts == 0 {
		ts = GetCurrentTS()
	}
	payload := map[string]interface{}{
		"alert": map[string]interface{}{
			"data":       alertData,
			"message":    alertMessage,
			"alert_type": alertLevel,
			"alert_key":  alertKey,
			"timestamp":  ts,
		},
	}

	return payload
}