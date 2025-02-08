/*
 * The idea here is to represent the state of each digital pin
 * using a bit. We have 11 pins so we will use a 16 byte type.
 * The pin to bit mapping will be as below from low to high byte
 * high bits ...    low bits
 * 0 0 0 0    0 0   0   0     0  0  0  0      0  0  0  0
 *              SD3 D10 D9    D8 D7 D6 D5     D3 D2 D1 D0
 *
 * AND mask for each pin will be
 *  D0  -> 0x1    = 1
 *  D1  -> 0x2    = 2
 *  D2  -> 0x4    = 4
 *  D3  -> 0x8    = 8
 *  D5  -> 0x10   = 16
 *  D6  -> 0x20   = 32
 *  D7  -> 0x40   = 64
 *  D8  -> 0x80   = 128
 *  D9  -> 0x100  = 256
 *  D10 -> 0x200  = 512
 *  SD3 -> 0x400  = 1024
 *
 *  We will use 32 bits structure with
 *  lower 2 bytes for PIN ON states
 *  and higher 2 bytes for PIN OFF states
 *
 * LIGHT VALUES
 * 101
 *		               0100 0000
 * 102
 * 0100 0000 0000 0000 0000 0000
 *
 *
 * FAN VALUES
 * 101 & 103 & 108  (ON & ONE & 1)
 * 0000 0000 0011 1100 0000 0000 0000 0010 (3932162)
 *
 * 102 (OFF)
 * 0000 0000 0011 1110 0000 0000 0000 0000 (4063232)
 *
 * 104 & 109 (TWO & 2)
 * 0000 0000 0011 1010 0000 0000 0000 0100 (3801092)
 *
 * 105 & 110 (THREE & 3)
 * 0000 0000 0011 0110 0000 0000 0000 1000 (3538952)
 *
 * 106 & 111 (FOUR & 4)
 * 0000 0000 0010 1110 0000 0000 0001 0000 (3014672)
 *
 * 107 & 112 & 113 (FIVE & 5)
 * 0000 0000 0001 1110 0000 0000 0010 0000 (1966112)
 *
 */

package entityset

import (
	"encoding/json"
	"hombot/errors"
	"hombot/logging"
	"hombot/utils/constants"

	//"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

const NOT_FOUND int = -1

//list of value identifiers
// const VALUE_ON_ID uint16 = 100
// const VALUE_OFF_ID uint16 = 101

func init() {

}

// ======================================================
type DeviceInfo struct {
	DevceId string
	Command string
	Value   uint32
}

type Values struct {
	ValueList []string
	Ordered   bool
}

type Mappings struct {
	P              uint8
	DeviceMappings map[uint16]DeviceInfo
}

// ======================================================
type EntitySet struct {
	ValueIds       map[uint16]Values
	EntityMappings map[string]map[uint16]Mappings
	Plugins        map[string]string
}

var logger *logging.Logging = nil

var gEntitySet *EntitySet = nil
var mutex sync.Mutex

func GetEntitySet() *EntitySet {
	mutex.Lock()
	if gEntitySet == nil {
		gEntitySet = new(EntitySet)
	}
	mutex.Unlock()
	return gEntitySet
}

func (e *EntitySet) Init() int {
	var entityFile string

	if logger != nil {
		return errors.SUCCESS
	}
	//get logger first
	logger, _ = logging.GetLogger("", 0)

	goBinPath := os.Getenv("GOBIN")
	entityFile = path.Join(goBinPath, constants.ENTITY_FILE)
	logger.Info("Opening entities file %s", entityFile)

	jsonData, err := os.ReadFile(entityFile)
	if err != nil {
		logger.Error("Error opening file: %v", err)
		return errors.ENTITYSET_FAILED_READ_ENTITIES_FILE
	}

	logger.Info("Reading defined entities from %s", entityFile)
	err = json.Unmarshal(jsonData, &e)
	if err != nil {
		logger.Error("Error unmarshaling entites: %v", err)
	}

	logger.Debug("Loaded entities. ValueIds=%v", e.ValueIds)
	logger.Debug("Mappings=%v", e.EntityMappings)
	logger.Debug("Plugins=%v", e.Plugins)
	return errors.SUCCESS
}

/*
 * Searches for the listed entities like fan, light, etc.
 * in the supplied string
 */
func (e *EntitySet) SearchEntity(refString string) (string, int) {
	var pos int = NOT_FOUND
	var keyFound string

	logger.Debug("SearchEntity: Mappings=%v", e.EntityMappings)

	for key, _ := range e.EntityMappings {
		logger.Debug("searching entity [%s] in refstring [%s]", key, refString)
		pos = e.findValInString(refString, 0, key)
		if pos == NOT_FOUND {
			logger.Debug("entity [%s] Not Found", key)
			continue
		} else {
			logger.Info("entity [%s] Found", key)
			keyFound = key
			break
		}
	}
	if keyFound == "" {
		return "", errors.ENTITY_KEY_NOT_FOUND_REF_STRING
	}
	return keyFound, errors.SUCCESS
}

/*
 *	Returns the Mappings defined for the entity
 */
func (e *EntitySet) GetEntityMappings(entity_name string) *map[uint16]Mappings {
	val, ok := e.EntityMappings[entity_name]
	if !ok {
		return nil
	}

	return &val
}

/*
 *	Returns the Plugin for the entity
 */
func (e *EntitySet) GetEntityPlugin(entity_name string) string {
	val, ok := e.Plugins[entity_name]
	if !ok {
		return ""
	}
	return val
}

/*
 *	Returns the Values defined for the entity
 */
func (e *EntitySet) GetEntity(entity_name string) (*map[uint16]Values, *map[uint16]Mappings) {
	valuesMap := make(map[uint16]Values)
	val, ok := e.EntityMappings[entity_name]
	if !ok {
		return nil, nil
	} else {
		logger.Debug("EntityMappings is [%v] for [%v]", val, entity_name)
		for valId := range val {
			logger.Debug("valId: [%v] ", valId)
			vals, ok := e.ValueIds[valId]
			if ok {
				valuesMap[valId] = vals
			}
		}
	}
	return &valuesMap, &val
}

/*
 * Searches for the value sets in the reference string and
 * returns the valueset id with 0 as error code if found
 * or 0 with an error code if not found
 */
func (e *EntitySet) SearchValues(refString string, valueSet *map[uint16]Values, mappings *map[uint16]Mappings) (valueSetId uint16, errCode int) {
	var valIndex int8 = 0
	var findPos int = 0
	var keyFoundCnt int = 0
	var valCnt int = 0

	//var total_keys int = len(*valueSet)
	var priorityValueId uint16 = 0
	var lastPriority uint8 = 0
	logger.Info("Searching values in Reference string is %s", refString)
	for key, values := range *valueSet {
		logger.Debug("Search valueSet is [%v]", values.ValueList)
		valCnt = len(values.ValueList)

		//reset the counters
		keyFoundCnt = 0
		findPos = 0

		//ValueList here is an array of values
		//these values will be searched in order
		for _, val := range values.ValueList {
			logger.Debug("Finding value %s in refString %v", val, refString)
			findPos = e.findValInString(refString, findPos, val)
			if findPos == NOT_FOUND {
				break
			} else {
				keyFoundCnt++
				logger.Debug("Found %v", val)
			}
			findPos += len(val)
		}
		if keyFoundCnt == valCnt {
			//get the mappings struct to find the priority for this entity
			keyMapping := (*mappings)[key]
			if keyMapping.P > lastPriority {
				lastPriority = keyMapping.P
				priorityValueId = key
			}
			//return key, errors.SUCCESS
		}
		valIndex++
	}
	if priorityValueId > 0 {
		return priorityValueId, errors.SUCCESS
	} else {
		return 0, errors.ENTITY_VALUE_NOT_FOUND_REF_STRING
	}
}

// ======================================================
// ******* PRIVATE *********
func (e *EntitySet) findValInString(refString string, pos int, val string) int {
	var searchpos int = NOT_FOUND
	var valLen int = len(val)

	refstr := refString[pos:]

	searchpos = strings.Index(refstr, val)
	if searchpos >= 0 {
		x := searchpos + valLen
		//check for space as previous char
		if (searchpos > 0) && (refstr[searchpos-1:searchpos] != " ") {
			searchpos = NOT_FOUND
		} else if x < len(refstr) {
			//check for space as next char
			if refstr[x:x+1] != " " {
				searchpos = NOT_FOUND
			}
		}
	}

	return searchpos
}
