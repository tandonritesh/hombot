package speech

import (
	"fmt"
	errors "hombot/errors"
	"hombot/logging"
	"hombot/scheduler"
	"hombot/utils/constants"
	"io"
	"log"
	"time"

	speech "cloud.google.com/go/speech/apiv1p1beta1"
	"golang.org/x/net/context"

	//speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1p1beta1"
	speechpb "cloud.google.com/go/speech/apiv1p1beta1/speechpb"
)

const AUDIORATE_HERTZ = 16000
const SINGLEUTTERANCE = true
const INTERIMRESULTS = false
const AUDIO_ENCODING = speechpb.RecognitionConfig_LINEAR16

var logger *logging.Logging
var dataChannel chan string
var dataBuf []byte

func GetDataChannel() chan string {
	return dataChannel
}

type SpeechClient = speech.Client

type SpeechClientArgs struct {
	ss speechpb.Speech_StreamingRecognizeClient
	sc *speech.Client
}

func sendAudio(speechStream speechpb.Speech_StreamingRecognizeClient, micBuf *MicAudioStream, quitch chan bool) {
	//read the incoming voice data
	//buffer creation
	var stopsend bool = false
	var streamRequest speechpb.StreamingRecognizeRequest
	var audioContent speechpb.StreamingRecognizeRequest_AudioContent

	logger.Info("Starting reading buffer for Speech recognition")
	for {
		select {
		case send := <-quitch:
			logger.Info("now stopping = %v", send)
			stopsend = send
			break
		default:
			//logger.Printf("in case default")
			break
		}
		if stopsend {
			logger.Info("Now Quiting")
			return
		}

		//logger.Info("Reading MicBuffer")
		dataBuf, nBytes, speech_err := micBuf.Read(dataBuf)
		if speech_err != nil {
			if speech_err == io.EOF {
				//logger.Info("Speech buffer is empty")
			} else {
				logger.Error("Error reading data from Mic buffer. Err=%v", speech_err)
				//we can close the input stream
				err := speechStream.CloseSend()
				if err != nil {
					logger.Error("Error closing the speech stream. Err=%v", err)
				}
			}
		}
		// if nBytes > 0 {
		// 	logger.Info("Read MicBuf len: %v, nBytes: %v", len(dataBuf), nBytes)
		// }

		//nBytes, err = os.Stdin.Read(dataBuf)
		// if err == io.EOF {
		// 	//we can close the input stream
		// 	err := speechStream.CloseSend()
		// 	if err != nil {
		// 		logger.Error("Error closing the speech stream. Err=%v", err)
		// 	}
		// 	return
		// }

		// if err != nil {
		// 	logger.Error("Failed to read data from Stdin")
		// 	continue
		// }

		if !stopsend {
			//we have some data on the Stdin, lets send it to speech stream
			if nBytes > 0 {
				audioContent.AudioContent = dataBuf[:nBytes]
				streamRequest.StreamingRequest = &audioContent

				//logger.Info("Sending speech data")
				err := speechStream.Send(&streamRequest)
				if err != nil {
					logger.Error("There was an error sending audio content. %v", err)
				}

				//logger.Info("Sent SpeechData %v", nBytes)
			}
		}
	}
}

func Destroy(client *SpeechClient) {
	client.Close()
	logging.Destroy()

}

func InitClient() (*SpeechClient, int) {
	var errCode int
	var client *speech.Client

	//configure logger first
	logger, errCode = logging.GetLogger("", 0)
	//defer logging.Destroy()
	if errCode != errors.SUCCESS {
		log.Printf("Failed to get logger. errCode: %d", errCode)
		return nil, errors.SPEECH_CLIENT_FAILED_INIT_LOGGER
	}

	logger.Info("Starting Speech Recognition for home automation")

	client, err := speech.NewClient(context.Background())
	if err != nil {
		logger.Error("Failed to create speech client with error %v", err)
		return nil, errors.SPEECH_CLIENT_CREATE_ERROR
	} else {
		logger.Info("Speech Client created successfully")
	}

	dataChannel = make(chan string)
	logger.Info("Data Channel created successfully")

	dataBuf = make([]byte, 4096)

	return client, errors.SUCCESS
}

func InitDetector(client *SpeechClient, micBuffer *MicAudioStream) (string, int) {
	logger.Info("Speech InitDetector Called")
	var speechText string
	var speechStream speechpb.Speech_StreamingRecognizeClient

	ctx := context.Background()

	speechStream, err := client.StreamingRecognize(ctx)
	if err != nil {
		logger.Error("Failed to create speech stream. %v", err)
		return speechText, errors.SPEECH_CLIENT_STREAM_CREATE_ERROR
	} else {
		logger.Info("Speech Stream created successfully")
	}

	logger.Info("Configuring Speech Stream")
	var streamRequest speechpb.StreamingRecognizeRequest
	var reqStreamingConfig speechpb.StreamingRecognizeRequest_StreamingConfig
	var streamConfig speechpb.StreamingRecognitionConfig
	var config speechpb.RecognitionConfig
	var metadata speechpb.RecognitionMetadata

	var spCtx speechpb.SpeechContext
	spCtx.Phrases = []string{"fan", "fan on", "on fan", "fan off", "off fan",
		"speed 1", "speed 2", "speed 3", "speed 4", "speed 5",
		"speed one", "speed two", "speed three", "speed four", "speed five",
		"light", "lights", "on", "off", "bulb", "tubelight",
		"balcony light on", "balcony light off",
		"washing area light on", "washing area light off"}

	//metadata.InteractionType = speechpb.RecognitionMetadata_VOICE_COMMAND
	metadata.MicrophoneDistance = speechpb.RecognitionMetadata_MIDFIELD
	metadata.OriginalMediaType = speechpb.RecognitionMetadata_AUDIO
	metadata.RecordingDeviceType = speechpb.RecognitionMetadata_OTHER_INDOOR_DEVICE

	config.Encoding = AUDIO_ENCODING
	config.SampleRateHertz = AUDIORATE_HERTZ
	config.LanguageCode = "en-IN"
	config.SpeechContexts = []*speechpb.SpeechContext{&spCtx}
	//config.Model = "command_and_search"
	config.Metadata = &metadata
	//config.EnableSpeakerDiarization = true

	streamConfig.Config = &config
	streamConfig.SingleUtterance = SINGLEUTTERANCE
	streamConfig.InterimResults = INTERIMRESULTS
	reqStreamingConfig.StreamingConfig = &streamConfig
	streamRequest.StreamingRequest = &reqStreamingConfig

	logger.Info("\nStream Request: %v\n", streamRequest.GetStreamingConfig())
	err = speechStream.Send(&streamRequest)
	if err != nil {
		logger.Error("Failed to configure speech stream. %v", err)
		return speechText, errors.SPEECH_CLIENT_STREAM_CONFIG_ERROR
	} else {
		logger.Info("Speech Stream configured successfully")
	}
	defer speechStream.CloseSend()

	quitchan := make(chan bool, 1)
	logger.Info("Channel created")

	//Starting audio content routine
	logger.Info("Starting Audio Content routine")
	go sendAudio(speechStream, micBuffer, quitchan)
	logger.Info("Audio Content routine started")

	logger.Info("Entering loop to read response")
	//we will wait for 3 seconds after receiving the speech end
	//event before closing the streams and quiting the speech component
	var hasSpeechEnded bool = false
	var clientClosed bool = false

	for !clientClosed {
		if hasSpeechEnded == true {
			var tmr scheduler.ITimer = scheduler.NewSchdTimer()
			tmr.SetSeconds(3)
			var args SpeechClientArgs
			args.ss = speechStream
			args.sc = client
			var errCode int = tmr.Start(&args, stopSpeechClient)
			logger.Debug("Started the timer")
			if errCode != errors.SUCCESS {
				logger.Debug("Timer failed: errCode: %v", errCode)
				break
			}
		}
		//this is a blocking call till speech component
		//returns some data here
		logger.Info("Waiting for Speech Response")
		resp, err := speechStream.Recv()
		logger.Info("Response received %v", resp)
		fmt.Printf("Response received\n")
		if err == io.EOF {
			logger.Info("Response was EOF")
			quitchan <- true
			logger.Info("Breaking Speech Loop")
			break
		}
		if err != nil {
			logger.Error("Failure streaming results: %v", err)
			quitchan <- true
			logger.Info("Returning from Speech Loop")
			return "", errors.SPEECH_CLIENT_RECV_API_ERR
		}

		if err := resp.Error; err != nil {
			logger.Error("Failiure recognizing audio: %v", err)
			hasSpeechEnded = true
			quitchan <- true
			continue
		}
		//speech detection ended, lets not send any more data
		if resp.SpeechEventType == speechpb.StreamingRecognizeResponse_END_OF_SINGLE_UTTERANCE {
			logger.Info("Ending Speech")
			hasSpeechEnded = true
			quitchan <- true
		}

		var confidence float32

		for _, result := range resp.Results {
			var alternative []*speechpb.SpeechRecognitionAlternative = result.GetAlternatives()

			logger.Debug("Total alternatives %d", len(alternative))
			logger.Debug("Result: %+v, confidence %v\n", alternative[0].GetTranscript(), confidence)

			if result.IsFinal {
				confidence = alternative[0].GetConfidence()
				speechText = alternative[0].GetTranscript()
				logger.Info("Final Result: %+v, confidence %v\n", speechText, confidence)
				fmt.Printf("Final Result: %+v, confidence %v\n", speechText, confidence)
				//dataChannel <- data
				//printSpeechData(speechText, confidence)
				speechText = constants.SPEECH_DATA_HEADER + speechText
				speechStream.CloseSend()
				clientClosed = true
				break
			}
		}
	}
	logger.Info("Returning from Init Detector SpTxt: %v", speechText)
	// [END speech_streaming_mic_recognize]
	return speechText, errors.SUCCESS
}

// func main() {
// 	ret := InitDetector()
// 	if ret != errors.SUCCESS {
// 		if ret == errors.SPEECH_CLIENT_FAILED_INIT_LOGGER {
// 			log.Panicf("There was error initializing speech detector: code %d", ret)
// 		} else {
// 			logger.Panicf("There was error initializing speech detector: code %d", ret)
// 		}
// 	}
// }

func stopSpeechClient(tm time.Time, args interface{}) {
	if scArgs, ok := args.(*SpeechClientArgs); ok {
		scArgs.ss.CloseSend()
		//scArgs.sc.Close()
	} else {
		logger.Error("Received invalid args type from scheduler Timer  ")
	}

}

func printSpeechData(data string, confidence float32) {
	if confidence > constants.MIN_CONFIDENCE {
		//fmt.Printf("%s%s\n", constants.SPEECH_DATA_HEADER, data)
		fmt.Printf("Speech Data %s:%s", constants.SPEECH_DATA_HEADER, data)
	} else {
		//fmt.Printf("%s%s\n", constants.SPEECH_DATA_HEADER, constants.LOW_CONFIDENCE)
		fmt.Printf("%s%s", "", constants.LOW_CONFIDENCE)
	}
}
