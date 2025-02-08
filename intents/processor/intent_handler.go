package IntentHandler

import (
	"fmt"
	"hombot/errors"
	executorfactory "hombot/intentexecutors"
	"hombot/intents"
	"hombot/intents/entityset"
	"hombot/logging"
	"hombot/utils/constants"
	"hombot/utils/ipcpipe"
	"strings"
)

var entities *entityset.EntitySet

var logger *logging.Logging

func Init() int {
	var errCode int

	//get logger first
	logger, _ = logging.GetLogger("", 0)

	intents.SetLogger(logger)

	//intialize the executor factory
	errCode = executorfactory.Init()
	if errCode != errors.SUCCESS {
		return errCode
	}

	entities = entityset.GetEntitySet()

	//initialize the entity set
	errCode = entities.Init()
	if errCode != errors.SUCCESS {
		return errCode
	}

	errCode = ipcpipe.Init(constants.SERVER, dataCallback)
	if errCode != errors.SUCCESS {
		logger.Error("Failed to initialize IPC with err: %d", errCode)
		return errCode
	}

	return errors.SUCCESS
}

func Destroy() {
	executorfactory.Destroy()
}

func dataCallback(buf []byte) {
	//logger.Info("receieved data")
	//logger.Info("receieved data Len:%d", len(buf))
	logger.Info("intent data receieved  %d: %v", len(buf), string(buf))
	errCode := getIntent(buf)
	logger.Info("getIntent returned %d", errCode)
}

func getIntent(buf []byte) int {

	//we got the intent buffer now convert it back to struct
	var refString string
	var intentBuf intents.IntentBuffer

	intentBuf.FromBytes(buf)
	logger.Debug("IntentBuf: %v", intentBuf)

	refString = strings.ToLower(intentBuf.GetRefString())

	intentId, errCode := entities.SearchEntity(refString)
	if errCode != errors.SUCCESS {
		logger.Error("Search entity failed with %d", errCode)
		return errCode
	}
	logger.Debug("Intent Id: %v", intentId)

	values, mappings := entities.GetEntity(intentId)
	if values == nil {
		return errors.ENTITY_NOT_FOUND_IN_ENTITYSET
	}
	logger.Info("Values: %v, Mappings: %v", values, mappings)

	valueSetId, errCode := entities.SearchValues(refString, values, mappings)
	if errCode != errors.SUCCESS {
		logger.Error("Failed to get values for entity %s", intentId)
		return errCode
	}
	logger.Debug("Found ValueSet ID = %d", valueSetId)

	//we got the entity and the matching valueset
	var intent intents.Intent

	intent.SetSourceId(intentBuf.GetSourceId())
	intent.SetAddr(intentBuf.GetAddr())
	intent.SetKey(intentId)
	intent.SetValueId(valueSetId)
	intent.SetRefString(refString)

	fmt.Printf("Intent: %d, %s, %s, %d \n", intent.GetSourceId(), intent.GetAddr(), intent.GetKey(), intent.GetValueId())
	logger.Info("Intent: %d, %s, %s, %d ", intent.GetSourceId(), intent.GetAddr(), intent.GetKey(), intent.GetValueId())
	//get the executor for this intent
	fExecute, etype, errCode := executorfactory.GetExecutor(&intent)
	if errCode != errors.SUCCESS {
		logger.Error("Failed to Get executor for %s. errCode: %d", intent.GetKey(), errCode)
		return errCode
	}
	logger.Info("Got executor %s for intent %s", etype, intent.GetKey())

	//load the intent executor
	// executor, errCode := executorfactory.LoadExecutor(executorPath)
	// if errCode != errors.SUCCESS {
	// 	logger.Error("Failed to load executor %s %s. errCode: %d", etype, executorPath, errCode)
	// 	return errCode
	// }
	// logger.Info("Executor %s loaded", etype)

	//get the executor interface pointers
	// fInit, fExecute, _, errCode := executorfactory.GetFunctions(executor)
	// if errCode != errors.SUCCESS {
	// 	logger.Error("Failed to get executor interface. errCode", errCode)
	// 	return errCode
	// }
	// logger.Info("Got the executor interface")

	//execute the executor interface in sequence
	//defer fDestroy.(func())()

	// errCode = fInit.(func() int)()
	// logger.Info("Executor %s successfully initialzed", etype)

	errCode = fExecute.(func(*intents.Intent) int)(&intent)
	if errCode != errors.SUCCESS {
		logger.Error("Failed to execute executor. errCode: %v", errCode)
		return errCode
	}

	return errors.SUCCESS
}
