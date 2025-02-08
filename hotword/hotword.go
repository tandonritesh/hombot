package hotword

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hombot/errors"
	"hombot/logging"
	"hombot/speech"
	"hombot/utils"

	. "github.com/Picovoice/porcupine/binding/go/v3"
	"github.com/gordonklaus/portaudio"
	"github.com/stianeikeland/go-rpio"
)

var READY_LED = 17
var LISTEN_LED = 18

var dataChannel chan []byte

// var detector snowboydetect.SnowboyDetect
var porcupine Porcupine

var logger *logging.Logging

var pPortAudioStream *portaudio.Stream

// buffer to read data using portaudio
var audioInBuf []int16

// buffer to write portaudio data into LE
var dataBuf *bytes.Buffer

var micAudioStream *speech.MicAudioStream

var speechClient *speech.SpeechClient

func GetDataChannel() chan []byte {
	return dataChannel
}

func writeBytes(sAS *speech.MicAudioStream, quitch chan bool, start chan bool) {
	// buf := make([]byte, 2048)
	logger.Info("Speech write to MicAudioStream buffer initialized")
	var quitAudioRead bool = false
	var err error
	var errCode int
	var startAudioRead bool = false
	// // buffer to write portaudio data into LE
	// var dataBuf *bytes.Buffer
	// dataBuf = new(bytes.Buffer)

	//do a blocking call to the channel. Start only when asked for
	for !startAudioRead {
		startAudioRead = <-start
		logger.Info("Reading speech data")
		fmt.Printf("Reading speech data\n")
		for {
			select {
			case <-quitch:
				quitAudioRead = true
				break
			default:
				break
			}
			if quitAudioRead {
				logger.Info("Quiting Reading microphone data")
				startAudioRead = false
				quitAudioRead = false
				break
			}
			//fmt.Printf("\nNow Reading\n")
			// _, err = os.Stdin.Read(buf)
			// if err == io.EOF {
			// 	logger.Error("It is EOF")
			// 	break
			// }
			dataBuf.Reset()
			_, errCode = getAudioData(dataBuf)
			if errCode != errors.SUCCESS {
				logger.Error("1 - Failed to read Audio data. errCode: %d", errCode)
				break
			}

			//logger.Info("Writing speech data %v", dataBuf.Len())
			_, err = sAS.Write(*dataBuf)
			//_, err = stdinpipe.Write(dataBuf.Bytes())
			if err != nil {
				logger.Error("\nError writing speech bytes: %v\n", err)
				continue
			}
		}
	}
}

// func read(detector snowboydetect.SnowboyDetect, sd string) {
func read(detector Porcupine, spchAudioStream *speech.MicAudioStream) {
	var hotWordDetect bool = false
	var errCode int
	var err error
	//var stdinpipe io.WriteCloser
	var wakewordBuf []int16
	// buffer to write portaudio data into LE
	// var dataBuf *bytes.Buffer
	// dataBuf = new(bytes.Buffer)
	var speechResponse string

	quitch := make(chan bool, 1)
	start := make(chan bool)

	logger.Info("List of Devices")
	devInfo, err := portaudio.Devices()
	if err != nil {
		logger.Error("Portaudio Error: %v", err)
	}
	logger.Info("Devices: %v", devInfo)

	logger.Info("Default Input Device")
	defInDev, err := portaudio.DefaultInputDevice()
	if err != nil {
		logger.Error("Portaudio Error: %v", err)
	}
	logger.Info("Default In Device: %v", defInDev)

	logger.Info("Opening PortAudio stream")
	defer portaudio.Terminate()
	pPortAudioStream, err = portaudio.OpenDefaultStream(1, 0, 16000, len(audioInBuf), audioInBuf)
	if err != nil {
		logger.Panicf("Failed to open default stream %v", err)
		return
	}
	defer pPortAudioStream.Close()
	logger.Info("Successfully opened default stream for portaudio")

	logger.Info("Starting PortAudio stream")
	err = pPortAudioStream.Start()
	if err != nil {
		logger.Panicf("Failed to initialize portaudio %v", err)
		return
	}

	//start speech loop
	go writeBytes(spchAudioStream, quitch, start)

	logger.Info("Started Reading for hotword")
	fmt.Println("Started Reading for hotword")
	utils.SetPinState(READY_LED, rpio.High)
	utils.SetPinState(LISTEN_LED, rpio.Low)
	for {
		wakewordBuf, errCode = getAudioData_int16()
		if errCode != errors.SUCCESS {
			logger.Error("0 - Failed to read Audio data. errCode: %d", errCode)
			continue
		}
		//logger.Info("Received Hotword data: %d", wakewordBuf)

		if !hotWordDetect {
			// logger.Debug("Calling Detect with datasize %v", dataBuf.Len())
			hotWordDetect = detect(detector, wakewordBuf)
			if !hotWordDetect {
				continue
			} else {
				logger.Info("Hotword detected")
				fmt.Print("Hotword detected\n")
			}
		}
		hotWordDetect = false

		//turn on the speak LED
		utils.SetPinState(LISTEN_LED, rpio.High)
		utils.SetPinState(READY_LED, rpio.Low)

		//start the audio read from microphone
		logger.Info("starting MicRead")
		start <- true
		logger.Info("MicRead started")

		logger.Info("Starting speech read")
		speechResponse, errCode = speech.InitDetector(speechClient, micAudioStream)
		logger.Info("Speech InitDetector returned: %v, %v", speechResponse, errCode)
		if errCode != errors.SUCCESS {
			logger.Error("Failed to initialize speech detector. errCode: %d", errCode)
			logger.Info("Moving back to Re-Reading for Hotword")
			logger.Info("Speech module completed with failure")

			handleSpeechCompletion(quitch)
		} else {
			logger.Info("Speech module completed successfully")

			logger.Info("sending speech response %v", speechResponse)
			dataChannel <- []byte(speechResponse)
			logger.Info("Sent speech text from speech.dataChannel")

			handleSpeechCompletion(quitch)
		}

		logger.Info("Re-Reading for Hotword")
		fmt.Println("Waiting for Hotword")
	}

	//logger.Info("\nOut of Read\n")
}

func handleSpeechCompletion(quitch chan bool) {
	quitch <- true
	logger.Debug("quitch Set to TRUE")

	//turn off the speak LED
	utils.SetPinState(LISTEN_LED, rpio.Low)
	utils.SetPinState(READY_LED, rpio.High)

}

func detect(detector Porcupine, wwBuf []int16) bool {
	ret, err := detector.Process(wwBuf)
	//logger.Info("detector res: %d", ret)
	if err != nil {
		logger.Error("ERR_WAKE_WORD: %v", err)
	}

	if ret >= 0 {
		return true
	} else {
		return false
	}

}

// func detect(detector snowboydetect.SnowboyDetect, dat []byte, buflen int) bool {
// 	ptr := snowboydetect.SwigcptrInt16_t(unsafe.Pointer(&dat[0]))
// 	res := detector.RunDetection(ptr, buflen/2 /* len of int  */)
// 	if res == -2 {
// 		// logger.Debug("Snowboy detected silence")
// 	} else if res == -1 {
// 		logger.Error("Snowboy detection returned error")
// 	} else if res == 0 {
// 		// logger.Debug("Snowboy detected nothing")
// 	} else {
// 		logger.Debug("Snowboy detected keyword %d", res)
// 		return true
// 	}

// 	return false
// }

func Init(spchClient *speech.SpeechClient) int {
	//set the speech client
	speechClient = spchClient

	//get logger first
	logger, _ = logging.GetLogger("", 0)

	err := portaudio.Initialize()
	if err != nil {
		logger.Panicf("Failed to initialize portaudio %v", err)
		return -1
	}

	// Get the number of devices
	numDevices, err := portaudio.Devices()
	if err != nil {
		logger.Panicf("Failed to get audio devices %v", err)
		return -1
	}

	// List available devices
	for i, device := range numDevices {
		if device.MaxInputChannels > 0 {
			fmt.Printf("Device %d: %s\n", i, device.Name)
			fmt.Printf("  Input Channels: %d\n", device.MaxInputChannels)
			fmt.Printf("  Default Sample Rate: %f\n", device.DefaultSampleRate)
		}
	}

	//get the speech module path
	//speechDetectorPath := os.Args[1]

	//get a string list of models supplied
	//modelList = strings.Join(os.Args[2:], ",")
	//logger.Debug("listening for models %s", modelList)

	porcupine := Porcupine{
		AccessKey:       "UeoB96VUwdGO0zb63KhHfN79RsbhyMSUyZzaY5+T7adsoHjCUx4Prw==",
		BuiltInKeywords: []BuiltInKeyword{ALEXA, COMPUTER, OK_GOOGLE, HEY_GOOGLE}}
	//Sensitivities:   []float32{0.4, 0.4, 0.4, 0.4}}
	err = porcupine.Init()
	if err != nil {
		logger.Panicf("Failed to initialize Porcupine %v", err)
		return -1
	} else {
		logger.Info("Wakework detector created")
	}

	dataChannel = make(chan []byte)
	logger.Info("dataChannel created")

	// buffer to read data using portaudio
	audioInBuf = make([]int16, 512)
	// buffer to write portaudio data into LE
	dataBuf = new(bytes.Buffer)

	micAudioStream = speech.NewMicAudioStream()

	logger.Info("Starting hotword detection")
	go read(porcupine, micAudioStream)

	return 0
}

func getAudioData_int16() ([]int16, int) {
	err := pPortAudioStream.Read()
	if err != nil {
		logger.Error("Failed to read data from PortAudioStream %s", err)
		return nil, errors.HOTWORD_FAILED_PORTAUDIO_READ
	}

	return audioInBuf, 0
}

func getAudioData(buf *bytes.Buffer) (*bytes.Buffer, int) {
	err := pPortAudioStream.Read()
	if err != nil {
		logger.Error("Failed to read data from pPortAudioStream %s", err)
		return nil, errors.HOTWORD_FAILED_PORTAUDIO_READ
	}

	// write all the audio bytes read into dataBuf into LE format
	var cnt int = 0
	for _, v := range audioInBuf {
		cnt++
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			logger.Error("Failed to write audio to buffer. Err %v", err)
			return nil, errors.HOTWORD_FAILED_BUF_WRITE_FROM_PORTAUDIO
		}
	}

	return buf, 0
}

func Destroy() {
	fmt.Println("Destroying Hotword")
	//cportaudio.Terminate()
	utils.SetPinState(READY_LED, rpio.Low)
}
