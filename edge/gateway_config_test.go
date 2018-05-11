package edge

import (
	"fmt"
	"testing"
)

func TestMqttsGatewayConfig(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "mqtts"})

	if gw.Protocol != "mqtts" || gw.Host != "mqtt-broker.datonis.io" || gw.Port != 8883 {
		eGW := &GatewayConfig{Protocol: "mqtts", Host: "mqtt-broker.datonis.io", Port: 8883}
		t.Errorf("Expected Gateway to be %+v, but got %+v", eGW, gw)
	}
}

func TestMqttGatewayConfig(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "mqtt"})

	if gw.Protocol != "mqtt" || gw.Host != "mqtt-broker.datonis.io" || gw.Port != 1883 {
		eGW := &GatewayConfig{Protocol: "mqtt", Host: "mqtt-broker.datonis.io", Port: 1883}
		t.Errorf("Expected Gateway to be %+v, but got %+v", eGW, gw)
	}
}

func TestHttpsGatewayConfig(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "https"})

	if gw.Protocol != "https" || gw.Host != "api.datonis.io" {
		eGW := &GatewayConfig{Protocol: "https", Host: "api.datonis.io"}
		t.Errorf("Expected Gateway to be %+v, but got %+v", eGW, gw)
	}
}

func TestHttpGatewayConfig(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "http"})

	if gw.Protocol != "http" || gw.Host != "api.datonis.io" {
		eGW := &GatewayConfig{Protocol: "http", Host: "api.datonis.io"}
		t.Errorf("Expected Gateway to be %+v, but got %+v", eGW, gw)
	}
}

func TestUrlForMqttsRequest(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "mqtts"})
	url := gw.Url()
	expectedUrl := fmt.Sprintf("%s://%s:%d", "ssl", gw.Host, gw.Port)

	if url != expectedUrl {
		t.Errorf("Expected Gateway url to be %s, but got %s", expectedUrl, url)
	}
}

func TestUrlForMqttRequest(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "mqtt"})
	url := gw.Url()
	expectedUrl := fmt.Sprintf("%s://%s:%d", "tcp", gw.Host, gw.Port)

	if url != expectedUrl {
		t.Errorf("Expected Gateway url to be %s, but got %s", expectedUrl, url)
	}
}

func TestUrlForHttpsRequest(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "https"})
	url := gw.Url()
	expectedUrl := fmt.Sprintf("%s://%s", "https", gw.Host)

	if url != expectedUrl {
		t.Errorf("Expected Gateway url to be %s, but got %s", expectedUrl, url)
	}
}

func TestUrlForHttpRequest(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "http"})
	url := gw.Url()
	expectedUrl := fmt.Sprintf("%s://%s", "http", gw.Host)

	if url != expectedUrl {
		t.Errorf("Expected Gateway url to be %s, but got %s", expectedUrl, url)
	}
}

func TestAddCertificate(t *testing.T) {
	gw := InitGatewayConfig(&GatewayConfig{Protocol: "mqtts"})
	cert := "testing"

	if gw.AddCertificate(cert) != nil || gw.cert != "" {
		t.Errorf("Expected Gateway certificate to be empty, but got %s", gw.cert)
	}
}
