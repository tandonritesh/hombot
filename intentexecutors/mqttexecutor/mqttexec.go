package main

import (
	"fmt"
	"hombot/errors"
	"hombot/intents"
	"hombot/intents/entityset"
	"hombot/logging"
	mqttapp "hombot/mqtt"

	//"strconv"
	"strings"
)

const iot_server = "tcp://127.0.0.1:1883"
const user = ""
const pass = ""

const PLUGIN_NAME = "MQTT"
const clientId = "a-hombot-SPEECH_APP_100"

//const iot_server = "tcp://21onz8.messaging.internetofthings.ibmcloud.com:1883"
//const clientId = "a:21onz8:SPEECH_APP_100"
//const user = "a-21onz8-rarwmxp5fx"
//const pass = "kP(v23agSebF-xSlg*"

var logger *logging.Logging
var entities *entityset.EntitySet

var topic string = "hombot/type/controller/id/con001/evt/{{command}}/fmt/bin"

func init() {
}

func main() {
	Init()
	defer mqttapp.Destroy()
}

func Init() (exType string, errCode int) {
	var errcode int

	logger, _ = logging.GetLogger("", 0)

	errcode = mqttapp.Init(iot_server, clientId, user, pass)
	if errcode != errors.SUCCESS {
		logger.Error("Failed to initialize MQTT with err: %d", errcode)
		return PLUGIN_NAME, errcode
	}

	entities = entityset.GetEntitySet()
	//initialize the entity set
	errCode = entities.Init()
	if errCode != errors.SUCCESS {
		return PLUGIN_NAME, errCode
	}

	return PLUGIN_NAME, errors.SUCCESS
}

func Destroy() {
	mqttapp.Destroy()
}

func Execute(intent *intents.Intent) int {
	logger.Info("calling execute with %v", intent)

	//get topic and (command+value) from intent
	//	sourceId, key, valueId
	entityMap := entities.GetEntityMappings(intent.GetKey())
	if entityMap == nil {
		logger.Error("Entity %s not found in entity set", intent.GetKey())
		return errors.MQTT_ENTITY_NOT_FOUND_IN_SET
	}
	logger.Debug("EntityMap: %v", entityMap)

	//get the valueMapmappings
	mappings, ok := (*entityMap)[intent.GetValueId()]
	if ok == false {
		return errors.MQTT_VALUE_ID_NOT_FOUND
	}
	logger.Debug("Mappings: %v", mappings)

	devInfo, ok := mappings.DeviceMappings[intent.GetSourceId()]
	if ok == false {
		return errors.MQTT_SOURCE_ID_NOT_FOUND
	}
	logger.Debug("DevInfo: %v", devInfo)
	// switch intent.GetKey() {
	// case "fan", "light":
	// 	command = "pcs"
	// 	value = intent.GetValueId()
	// default:
	// 	command = "UNK"
	// 	value = 0
	// }

	//var dataBuf *bytes.Buffer = new(bytes.Buffer)
	//utils.EncodeUint16(value, dataBuf)

	//mqttapp.PublishMessage(strings.Replace(topic, "{{command}}", devInfo.Command, 1), 0, strconv.FormatUint(uint64(devInfo.Value), 10))
	var pubTopic = strings.Replace(topic, "{{command}}", devInfo.Command, 1)
	fmt.Printf("Topic: %v, val: %v\n", pubTopic, devInfo.Value)
	logger.Info("Topic: %v, val: %v", pubTopic, devInfo.Value)
	mqttapp.PublishMessage(pubTopic, 0, devInfo.Value)

	return errors.SUCCESS
}
