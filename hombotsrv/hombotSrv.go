package main

import (
	"hombot/errors"
	IntentHandler "hombot/intents/processor"
	"hombot/logging"
	"log"
	"os"
	"os/signal"
)

var logger *logging.Logging

func main() {
	var errCode int

	//configure logger first
	logger, errCode = logging.GetLogger("/tmp/logfile_hbsrv.log", 0)
	defer logging.Destroy()
	if errCode != errors.SUCCESS {
		log.Printf("Failed to get logger. errCode: %d", errCode)
	}

	errCode = IntentHandler.Init()
	defer IntentHandler.Destroy()
	if errCode != errors.SUCCESS {
		logger.Panicf("Failed to initialize IntentHandler with errCode: %d", errCode)
	}
	logger.Info("Intent handler initialized successfully")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Waiting for Ctrl+C")
	<-stop
	logger.Info("Hombot server Terminated")
	log.Println("Hombot server Terminated")
}
