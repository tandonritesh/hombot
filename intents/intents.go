package intents

import (
	"bytes"
	"hombot/logging"
	"hombot/utils"
	"strings"
)

var logger *logging.Logging

func SetLogger(loger *logging.Logging) {
	logger = loger
}

type Intent struct {
	sourceId  uint16
	srcAddr   string
	key       string
	valueId   uint16
	refString string
}

func (i *Intent) GetKey() string {
	return i.key
}
func (i *Intent) SetKey(key string) {
	i.key = key
}
func (i *Intent) GetValueId() uint16 {
	return i.valueId
}
func (i *Intent) SetValueId(valId uint16) {
	i.valueId = valId
}
func (i *Intent) GetRefString() string {
	return i.refString
}
func (i *Intent) SetRefString(str string) {
	i.refString = str
}
func (i *Intent) GetSourceId() uint16 {
	return i.sourceId
}
func (i *Intent) SetSourceId(id uint16) {
	i.sourceId = id
}
func (i *Intent) GetAddr() string {
	return i.srcAddr
}
func (i *Intent) SetAddr(addr string) {
	i.srcAddr = addr
}

func (i *Intent) ToBytes(dataBuf *bytes.Buffer) *bytes.Buffer {
	//get the buffer for 16 bit value id
	utils.EncodeUint16(i.sourceId, dataBuf)
	utils.EncodeString(i.srcAddr, dataBuf)
	utils.EncodeString(i.key, dataBuf)
	utils.EncodeUint16(i.valueId, dataBuf)
	utils.EncodeString(i.refString, dataBuf)

	return dataBuf
}

func (i *Intent) FromBytes(buf []byte) uint16 {
	var startLen uint16 = 0
	var str string
	var val uint16
	var refStr string

	//read sourceId
	val, startLen = utils.DecodeUint16(buf, startLen)
	i.sourceId = val

	str, startLen = utils.DecodeString(buf, startLen)
	i.srcAddr = str

	str, startLen = utils.DecodeString(buf, startLen)
	i.key = str

	val, startLen = utils.DecodeUint16(buf, startLen)
	i.valueId = val

	refStr, startLen = utils.DecodeString(buf, startLen)
	i.refString = refStr

	return startLen
}

//========================================================================
//========================================================================
type IntentBuffer struct {
	sourceId  uint16
	addr      string
	refString string
}

func (ib *IntentBuffer) SetAddr(addr string) {
	ib.addr = addr
}

func (ib *IntentBuffer) GetAddr() string {
	return ib.addr
}

func (ib *IntentBuffer) SetRefString(str string) {
	ib.refString = str
}
func (ib *IntentBuffer) SetRefStringFromBytes(buf []byte) {
	var builder strings.Builder
	builder.Write(buf)
	ib.refString = builder.String()
	ib.refString = ib.refString
}

func (ib *IntentBuffer) GetRefString() string {
	return ib.refString
}

func (ib *IntentBuffer) SetSourceId(id uint16) {
	ib.sourceId = id
}

func (ib *IntentBuffer) GetSourceId() uint16 {
	return ib.sourceId
}

func (ib *IntentBuffer) ToBytes(dataBuf *bytes.Buffer) *bytes.Buffer {
	//get the buffer for 16 bit value id
	utils.EncodeUint16(ib.GetSourceId(), dataBuf)
	utils.EncodeString(ib.addr, dataBuf)
	utils.EncodeString(ib.refString, dataBuf)

	return dataBuf
}

func (ib *IntentBuffer) FromBytes(buf []byte) uint16 {
	var startLen uint16 = 0
	var str string
	var sourceId uint16
	var refString string

	//read source id
	// sourceId = binary.LittleEndian.Uint16(buf[startLen : startLen+2])
	// startLen += 2
	sourceId, startLen = utils.DecodeUint16(buf, startLen)
	ib.SetSourceId(sourceId)

	//read address len
	// var bufLen uint16 = binary.LittleEndian.Uint16(buf[startLen : startLen+2])
	// startLen += 2
	// //read bufLen from buffer now into entity key
	// strBuilder.Write(buf[startLen : startLen+bufLen])
	// startLen += bufLen //+ 1 //add 1 for the separator :
	if logger != nil {
		logger.Debug("Buf=%v and startLen=%d", buf, startLen)
	}
	str, startLen = utils.DecodeString(buf, startLen)
	ib.SetAddr(str)

	// bufLen = binary.LittleEndian.Uint16(buf[startLen : startLen+2])
	// startLen += 2
	refString, startLen = utils.DecodeString(buf, startLen)
	ib.SetRefString(refString)

	return startLen
}
