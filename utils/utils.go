package utils

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/stianeikeland/go-rpio"
)

func EncodeUint16(val uint16, dataBuf *bytes.Buffer) {
	var buf []byte = []byte{0, 0}

	binary.LittleEndian.PutUint16(buf[0:], val)
	dataBuf.Write(buf)

}

func EncodeString(str string, dataBuf *bytes.Buffer) {
	var lenBuf []byte = []byte{0, 0}

	binary.LittleEndian.PutUint16(lenBuf[0:], uint16(len(str)))

	dataBuf.Write(lenBuf)
	dataBuf.WriteString(str)
}

func EncodeBytes(buf []byte, dataBuf *bytes.Buffer) {
	var lenBuf []byte = []byte{0, 0}

	binary.LittleEndian.PutUint16(lenBuf[0:], uint16(len(buf)))

	dataBuf.Write(lenBuf)
	dataBuf.Write(buf)

}

func DecodeUint16(buf []byte, index uint16) (val uint16, nextIndex uint16) {
	//read source id
	valUint16 := binary.LittleEndian.Uint16(buf[index : index+2])
	index += 2
	return valUint16, index

}

func DecodeString(buf []byte, index uint16) (val string, nextIndex uint16) {
	var strBuilder strings.Builder

	//read address len
	bufLen, index := DecodeUint16(buf, index)

	//read bufLen from buffer now into entity key
	strBuilder.Write(buf[index : index+bufLen])
	index += bufLen
	return strBuilder.String(), index
}

func DecodeBytes(buf []byte, index uint16) (val []byte, nextIndex uint16) {
	bufLen, index := DecodeUint16(buf, index)

	retBuf := buf[index : index+bufLen]
	index += bufLen

	return retBuf, index
}

func SetPinState(pin_no int, state rpio.State) {
	var pin = rpio.Pin(pin_no)
	
	var pinState = pin.Read()
	if pinState != state {
		pin.Write(state)
	}
	//rpio.ReadPin
}

// type OrderedList struct {

// 	key interface{}
// 	val interface{}
// }

// func (o *OrderedList) Create(size int) *OrderedList {

// }
