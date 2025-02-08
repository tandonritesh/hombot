package main

import (
	"bytes"
	"fmt"
	"hombot/errors"
	"hombot/hotword"
	"hombot/intents"
	"hombot/logging"
	"hombot/speech"
	"hombot/utils/constants"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/stianeikeland/go-rpio"
)

const SPEECH_SOURCE_ID = 0x1000

var logger *logging.Logging
var SPEECH_DATA_HEADER_BYTES []byte

func main() {
	var errCode int
	var errHotword int
	var speechClient *speech.SpeechClient

	//configure logger first
	logger, errCode = logging.GetLogger("/tmp/logfile_hb.log", 0)
	defer logging.Destroy()
	if errCode != errors.SUCCESS {
		log.Panicf("Failed to get logger. errCode: %d", errCode)
	}
	// if len(os.Args) < 2 {
	// 	log.Printf("Not enough Arguments %v", os.Args)
	// 	log.Printf("Required Arguments <command> <path to google speech detection module> <comma separated list of hotword model files>")
	// 	log.Panicf("Failure starting application. errCode: %d", errors.HOMBOT_NOT_ENOUGH_ARGS)
	// 	return
	// }
	logger.Info("Initializing RPi pins I/O")
	err := rpio.Open()
	defer rpio.Close()
	if err != nil {
		logger.Panicf("Failure initializing RPi pins I/O %s", err)
	}
	logger.Debug("Seeting PIN18 to output")
	rpio.PinMode(rpio.Pin(hotword.READY_LED), rpio.Output)
	rpio.PinMode(rpio.Pin(hotword.LISTEN_LED), rpio.Output)

	logger.Info("Initializing Speech Client")
	speechClient, errCode = speech.InitClient()
	defer speech.Destroy(speechClient)
	if errCode != errors.SUCCESS {
		logger.Panicf("Failed to create the Speech Client. errCode: %d", errCode)
	}
	logger.Info("Speech Client successfully initialized")

	logger.Info("Initializing hotword detection for home automation")
	errHotword = hotword.Init(speechClient)
	defer hotword.Destroy()
	if errHotword != errors.SUCCESS {
		logger.Panicf("Hotword detection initialization failed with %d", errHotword)
	}
	logger.Info("Hotword detection successfully initialized")

	logger.Info("Initializing IPC Client: Connecting to Server")
	conn, errCode := InitIPCClient()
	if errCode != errors.SUCCESS {
		logger.Panicf("Failed to initialize the IPC Client. errCode: %d", errCode)
	}

	go sendAudioText(conn)

	//get the speech header bytes
	SPEECH_DATA_HEADER_BYTES = bytes.NewBufferString(constants.SPEECH_DATA_HEADER).Bytes()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	logger.Info("Waiting for Ctrl+C")
	fmt.Println("Waiting for Ctrl+C")
	sig := <-sigChannel
	logger.Info("Received %s: Hombot Terminated", sig)
	log.Println("\n\nHombot Stopped\n\n")
}

func InitIPCClient() (net.Conn, int) {
	var conn net.Conn
	var err error

	logger.Info("Starting Client...")
	conn, err = net.Dial("tcp4", "localhost:12345")
	if err != nil {
		logger.Error("Error connecting to server: %v", err)
		return nil, errors.HOMBOT_FAILED_TO_INIT_IPCCLIENT
	}
	logger.Info("Client connected: LocalAddr: %v, RemoteAddr: %v", conn.LocalAddr(), conn.RemoteAddr())
	return conn, errors.SUCCESS
}

func sendAudioText(conn net.Conn) {
	var byteWrite int
	var err error

	var dataBuf *bytes.Buffer = new(bytes.Buffer)

	for {
		dataBuf.Truncate(0)
		//wait for the audio text
		logger.Info("Waiting for speech text")
		var buf []byte = <-hotword.GetDataChannel()
		logger.Info("Speech text: %v", string(buf))
		//check if we got the speech data
		if bytes.HasPrefix(buf, SPEECH_DATA_HEADER_BYTES) {
			//we got the data, now encode it with local address
			//and send it to intent handler
			var intentBuf intents.IntentBuffer

			intentBuf.SetSourceId(SPEECH_SOURCE_ID)
			intentBuf.SetAddr(conn.LocalAddr().String())
			intentBuf.SetRefStringFromBytes(buf[constants.SPEECH_DATA_HEADER_LEN:])
			intentBuf.ToBytes(dataBuf)
			//write the bytes to socket
			byteWrite, err = conn.Write(dataBuf.Bytes())
			if err != nil {
				logger.Error("Error writing data to socket: %v", err)
			}
			logger.Info("Data written to socket: %v: %d", string(buf), byteWrite)
			fmt.Printf("Data written to socket: %v: %d\n", string(buf), byteWrite)
		} else {
			logger.Error("No data to write to socket: %s", buf)
		}
	}
}
