package mqttapp

import (
	"encoding/binary"
	"github.com/eclipse/paho.mqtt.golang"
	"hombot/errors"
	"hombot/logging"
	"time"
)

type Payload struct {
	topic string
	qos   byte
	data  uint32
}

const TOKEN_WAIT = 30 * time.Second

var publishChannel chan Payload
var quitChannel chan bool

var logger *logging.Logging
var client mqtt.Client

func Init(server string, clientId string, user string, pass string) int {
	var opt *mqtt.ClientOptions = nil
	var token mqtt.Token

	logger, _ = logging.GetLogger("", 0)

	opt = mqtt.NewClientOptions()
	opt.AddBroker(server)
	opt.SetClientID(clientId)
	opt.SetCleanSession(true)
	opt.SetDefaultPublishHandler(msgCallback)
	opt.SetKeepAlive(2 * time.Second)
	opt.SetUsername(user)
	opt.SetPassword(pass)

	//logger.Info("Client Options %v", opt.)
	client = mqtt.NewClient(opt)
	logger.Info("Client object created")

	logger.Info("Connecting to IoT server")
	token = client.Connect()
	var ret bool = token.WaitTimeout(TOKEN_WAIT)
	if ret == false {
		logger.Error("Connect Timeout")
		return errors.MQTT_CONNECT_TIMEOUT
	}
	tknerr := token.Error()
	if tknerr != nil {
		logger.Error("Connect Error. err: %s", tknerr.Error())
		return errors.MQTT_CONNECT_ERROR
	}
	logger.Info("Connected to IoT server successfully")

	publishChannel = make(chan Payload)
	logger.Info("Publish Channel created")

	quitChannel = make(chan bool)
	logger.Info("Quit Channel created")

	go publishData()
	logger.Info("Publish GoRoutine started")

	return errors.SUCCESS
}

func PublishMessage(topic string, qos byte, data uint32) {
	var payload *Payload = new(Payload)

	payload.topic = topic
	payload.qos = qos
	payload.data = data

	publishChannel <- *payload
}

func Destroy() {
	//shutdown go
	quitChannel <- true
	close(publishChannel)
	client.Disconnect(10 * 1000)

}

//=======================================================
func publishData() {
	var payload Payload
	var quitnow bool = false

	logger.Info("Starting Publish GoRoutine")
	for {
		select {
		case payload = <-publishChannel:
		case <-quitChannel:
			quitnow = true
		}
		if quitnow {
			logger.Info("Quiting Publish GoRoutine")
			break
		}
		logger.Info("publishing data %v", payload)
		//convert data to bytes
		var byteBuf []byte = []byte{0, 0, 0, 0}
		binary.BigEndian.PutUint32(byteBuf[0:], payload.data)
		token := client.Publish(payload.topic, payload.qos, false, byteBuf)
		ret := token.WaitTimeout(TOKEN_WAIT)
		if ret == false {
			logger.Error("Publish Timeout")
			continue
		}
		tknerr := token.Error()
		if tknerr != nil {
			logger.Error("Publish Error. err: %s", tknerr.Error())
			continue
		}
		logger.Info("Message published successfully")
	}
}

func msgCallback(client mqtt.Client, msg mqtt.Message) {

}
