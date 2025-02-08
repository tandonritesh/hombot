package executorfactory

import (
	"hombot/errors"
	"hombot/intents"
	"hombot/intents/entityset"
	"hombot/logging"
	"os"
	"path"
	"plugin"
	"sync"
)

// type IExecutor interface {
// 	Init() int
// 	Destroy()
// 	Execute(intent *intents.Intent) int
// }

type Executor struct {
	path   string
	name   string
	plugin *plugin.Plugin
	fI     plugin.Symbol
	fE     plugin.Symbol
	fD     plugin.Symbol
}

func (ex *Executor) SetPath(p string) {
	ex.path = p
}
func (ex *Executor) GetPath() string {
	return ex.path
}
func (ex *Executor) SetName(nm string) {
	ex.name = nm
}
func (ex *Executor) GetName() string {
	return ex.name
}
func (ex *Executor) SetPlugin(pl *plugin.Plugin) {
	ex.plugin = pl
}
func (ex *Executor) GetPlugin() *plugin.Plugin {
	return ex.plugin
}
func (ex *Executor) SetFInit(f plugin.Symbol) {
	ex.fI = f
}
func (ex *Executor) GetFInit() plugin.Symbol {
	return ex.fI
}
func (ex *Executor) SetFExecute(f plugin.Symbol) {
	ex.fE = f
}
func (ex *Executor) GetFExecute() plugin.Symbol {
	return ex.fE
}
func (ex *Executor) SetFDestroy(f plugin.Symbol) {
	ex.fD = f
}
func (ex *Executor) GetFDestroy() plugin.Symbol {
	return ex.fD
}

var logger *logging.Logging
var mapExecutor map[string]Executor
var mutex sync.Mutex
var entities *entityset.EntitySet

//****************************************************

func Init() int {
	mapExecutor = make(map[string]Executor)

	logger, _ = logging.GetLogger("", 0)

	entities = entityset.GetEntitySet()
	//initialize the entity set
	errCode := entities.Init()
	if errCode != errors.SUCCESS {
		return errCode
	}

	return errors.SUCCESS
}

func Destroy() {
	//loop thru all the executors and unload the plugins

}

func GetExecutor(intent *intents.Intent) (fExecute plugin.Symbol, exType string, errCode int) {
	var goBinPath string
	var pluginFileName string = ""
	var executor Executor

	pluginFileName = entities.GetEntityPlugin(intent.GetKey())
	if pluginFileName == "" {
		logger.Error("No plugins exist to handle intent %s", intent.GetKey())
		return nil, "", errors.EX_PLUGIN_NOT_FOUND
	}

	//lock the function
	mutex.Lock()
	defer mutex.Unlock()
	//search the executor in the map, if not found load it and add it
	executor, ok := mapExecutor[pluginFileName]
	if ok == true {
		if executor.GetPlugin() != nil {
			logger.Info("We found the plugin loaded")
			return executor.GetFExecute(), executor.GetName(), errors.SUCCESS
		}
	}

	goBinPath = os.Getenv("GOBIN")
	//get the executor plugin path
	exPath := path.Join(goBinPath, pluginFileName)

	//now load the executor plugin
	exPl, errCode := loadExecutor(exPath)
	if errCode != errors.SUCCESS {
		logger.Error("Failed to load executor %s. errCode: %d", exPath, errCode)
		return nil, "", errCode
	}

	//now get the functions
	fI, fE, fD, errCode := getFunctions(exPl)
	if errCode != errors.SUCCESS {
		logger.Error("Failed to get executor interface. errCode: %d", errCode)
		return nil, "", errCode
	}

	//now initialize the executor plugin
	exTyp, errCode := fI.(func() (string, int))()
	if errCode != errors.SUCCESS {
		logger.Error("Failed to Initialize the executor %s. errCode: %d", exTyp, errCode)
		return nil, exTyp, errCode
	}
	logger.Info("Executor %s successfully initialzed", exTyp)

	//create executor structure and add it to the map
	executor.SetPath(exPath)
	executor.SetName(exTyp)
	executor.SetPlugin(exPl)
	executor.SetFInit(fI)
	executor.SetFExecute(fE)
	executor.SetFDestroy(fD)

	mapExecutor[pluginFileName] = executor

	return fE, exTyp, errors.SUCCESS
}

func loadExecutor(path string) (*plugin.Plugin, int) {
	var executor *plugin.Plugin
	var err error

	executor, err = plugin.Open(path)
	if err != nil {
		logger.Error("Failed to open plugin. err: %v", err)
		return nil, errors.EX_FAILED_TO_LOAD_PLUGIN
	}

	return executor, errors.SUCCESS
}

func getFunctions(executor *plugin.Plugin) (fInit plugin.Symbol, fExecute plugin.Symbol, fDestroy plugin.Symbol, errCode int) {

	var fI plugin.Symbol = nil
	var fE plugin.Symbol = nil
	var fD plugin.Symbol = nil
	var err error

	fI, err = executor.Lookup("Init")
	if err != nil {
		logger.Error("Failed to lookup function Init. err: %v", err)
		return nil, nil, nil, errors.EX_FAILED_TO_LOOKUP_INIT_FUNCTION
	}

	fE, err = executor.Lookup("Execute")
	if err != nil {
		logger.Error("Failed to lookup function Execute. err: %v", err)
		return nil, nil, nil, errors.EX_FAILED_TO_LOOKUP_EXECUTE_FUNCTION
	}

	fD, err = executor.Lookup("Destroy")
	if err != nil {
		logger.Error("Failed to lookup function Destroy. err: %v", err)
		return nil, nil, nil, errors.EX_FAILED_TO_LOOKUP_DESTROY_FUNCTION
	}

	return fI, fE, fD, errors.SUCCESS
}
