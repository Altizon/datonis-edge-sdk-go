package edge

import "fmt"

type GatewayConfig struct {
	AccessKey string
	SecretKey string
	Protocol  string
	Host      string
	Port      int32
	cert      string
}

func InitGatewayConfig(c *GatewayConfig) *GatewayConfig {
	if c.Protocol == "mqtt" {
		if c.Host == "" {
			c.Host = "mqtt-broker.datonis.io"
		}
		if c.Port == 0 {
			c.Port = 1883
		}
	} else if c.Protocol == "mqtts" {
		if c.Host == "" {
			c.Host = "mqtt-broker.datonis.io"
		}
		if c.Port == 0 {
			c.Port = 8883
		}
	} else if c.Host == "" {
		c.Host = "api.datonis.io"
	}

	return c
}

func (c *GatewayConfig) Url() string {
	url := ""
	if c.Protocol == "mqtt" || c.Protocol == "mqtts" {
		protocol := "tcp"
		if c.Protocol == "mqtts" {
			protocol = "ssl"
		}
		url = fmt.Sprintf("%s://%s:%d", protocol, c.Host, c.Port)
	} else {
		if c.Protocol == "https" {
			url = "https://" + c.Host
		} else {
			url = "http://" + c.Host
		}
		if c.Port != 0 {
			url = fmt.Sprintf("%s:%d", url, c.Port)
		}
	}
	return url
}

func (c *GatewayConfig) AddCertificate(cert string) *GatewayConfig {
	return nil
}
