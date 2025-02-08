package logging

import (
	"fmt"
	"hombot/errors"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

const LEVEL_TRACE = uint8(1)
const LEVEL_INFO = uint8(2)
const LEVEL_ERROR = uint8(3)

type Logging struct {
	logFilePath string
	logFile     *os.File
	flags       int
	logger      *log.Logger
}

func (l *Logging) Debug(format string, v ...interface{}) {
	var str string
	if v != nil {
		str = fmt.Sprintf(format, v...)
	} else {
		str = format
	}
	file, line := GetDetails()
	l.logger.Printf("D %s %d %s", *file, line, str)
}

func (l *Logging) Info(format string, v ...interface{}) {
	var str string
	if v != nil {
		str = fmt.Sprintf(format, v...)
	} else {
		str = format
	}
	file, line := GetDetails()
	l.logger.Printf("I %s %d %s", *file, line, str)
}

func (l *Logging) Error(format string, v ...interface{}) {
	var str string
	if v != nil {
		str = fmt.Sprintf(format, v...)
	} else {
		str = format
	}
	file, line := GetDetails()
	l.logger.Printf("E %s %d %s", *file, line, str)
}

func (l *Logging) Panicf(format string, v ...interface{}) {
	var str string
	if v != nil {
		str = fmt.Sprintf(format, v...)
	} else {
		str = format
	}
	file, line := GetDetails()
	l.logger.Panicf("P %s %d %s", *file, line, str)
}

// func (l *Logging) Fatalf(logLevel uint8, format string, v ...interface{}) {
// 	l.logger.Fatalf(format, v)
// }

// func (l *Logging) Fatalln(logLevel uint8, v ...interface{}) {
// 	l.logger.Fatalln(v)
// }

var loging *Logging = nil
var mutex sync.Mutex

func GetLogger(logPath string, flags int) (*Logging, int) {
	var err error

	mutex.Lock()
	defer mutex.Unlock()
	if loging == nil {
		loging = new(Logging)

		if logPath == "" {
			logPath = "/tmp/logfile.log"
		}

		if flags == 0 {
			flags = (log.LstdFlags)
		}
		loging.logFilePath = logPath
		loging.flags = flags

		//now create the logger
		loging.logFile, err = os.OpenFile(
			loging.logFilePath,
			os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Printf("Failed to initialize Log: %v", err)
			return nil, errors.LOGGER_FAILED_CREATE_LOG_FILE
		}

		loging.logger = log.New(loging.logFile, "", loging.flags)
		loging.Info("****************************************************")
		loging.Info("********** S T A R T I N G   N E W   L O G *********")
		loging.Info("****************************************************")
	}

	return loging, errors.SUCCESS
}

func Destroy() {
	defer loging.logFile.Close()
}

//==================================================
func GetDetails() (*string, int) {
	_, file, line, success := runtime.Caller(2)
	if success == true {
		var filename string = ""
		pos := strings.LastIndexByte(file, '/')
		if pos != -1 {
			filename = file[pos+1:]
		}
		return &filename, line
	}
	return nil, 0

}
