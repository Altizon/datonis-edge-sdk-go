# datonis-edge-sdk-go
Go language version of the Datonis Edge SDK

Configuring the Agent
---------------------

Modify the sample.go file as follows:

1. Add appropriate access_key and secret_key from the downloded key_pair in GatewayConfig struct.
2. Add Key and Name of the thing whose data you want to send to Datonis.
3. Finally add the metrics name and its value in getData function. You can also set waypoints and send it to Datonis.
4. Data can be send using HTTP or MQTT protocol for which appropriate funtion should be used.

Implementing Edge Agent
------------------------

You can then run example as follows:

go run examples/sample.go

Acknowledgments
---------------

 * We would like to thank [Josh Software](http://www.joshsoftware.com/) for providing the go SDK with MQTTS protocol.