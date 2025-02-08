package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
)

type DeviceInfo struct {
	DevceId string
	Command string
	Value   uint32
}

type Values struct {
	ValueList      []string
	DeviceMappings map[uint16]DeviceInfo
}

func bufTest1() {

	var buf bytes.Buffer
	var byteBuf []byte
	byteBuf = make([]byte, 10)

	//buf.WriteString("ABCD")
	//log.Print(buf.Bytes())
	//buf.WriteString("EFG")
	//log.Print(buf.Bytes())

	buf.WriteString("ABCDEFGH")
	log.Print("Full buffer: ", buf.Bytes())
	buf.Reset()
	log.Print(buf.Read(byteBuf))

	log.Print(byteBuf)
	buf.WriteString("IJKL")
	log.Print(buf.Read(byteBuf))
	log.Print(byteBuf)

	log.Print(buf.Bytes())
}

func main() {
	bufTest1()
	return
	var entityMap map[string]map[uint16]Values

	m1 := make(map[string]map[uint16]Values)

	val := m1["abc"]
	log.Print(val)
	//entityMap = make(map[string]map[uint16]Values)

	// //create device mappings
	// deviceMappings := make(map[uint16]DeviceInfo)
	// //mapping 1
	// var devinfo DeviceInfo
	// devinfo.DevceId = "dev1"
	// devinfo.Command = "psc"
	// devinfo.Value = 0x6
	// deviceMappings[0x1000] = devinfo

	// //mapping 2
	// devinfo.DevceId = "dev2"
	// devinfo.Command = "psc"
	// devinfo.Value = 0x6
	// deviceMappings[0x1001] = devinfo

	// //create values
	// valueMap := make(map[uint16]Values)
	// var values Values

	// values.Devicemappings = deviceMappings
	// values.ValueList = []string{"on"}
	// valueMap[101] = values

	// values.ValueList = []string{"off"}
	// valueMap[102] = values

	// entityMap["light"] = valueMap

	// jsonStr, err := json.Marshal(entityMap)
	// if err != nil {
	// 	log.Printf("%s", err)
	// }

	// log.Printf("%s", jsonStr)

	jsonData, err := ioutil.ReadFile("../entities.json")
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return
	}

	err = json.Unmarshal(jsonData, &entityMap)
	if err != nil {
		log.Printf("Error %v", err)
	}

	valMap := entityMap["light"]
	values := valMap[101]

	log.Print(values.ValueList)
	devMap := values.DeviceMappings
	devInfo := devMap[0x1001]

	log.Print(devInfo)

	log.Print("helo")
}
