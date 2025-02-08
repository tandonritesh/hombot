package errors

const SUCCESS = 0

// basic client errors 1001 - 1100
const SPEECH_CLIENT_ERROR_BASE = 1000
const SPEECH_CLIENT_CREATE_ERROR = SPEECH_CLIENT_ERROR_BASE + 1
const SPEECH_CLIENT_STREAM_CREATE_ERROR = SPEECH_CLIENT_ERROR_BASE + 2
const SPEECH_CLIENT_STREAM_CONFIG_ERROR = SPEECH_CLIENT_ERROR_BASE + 3
const SPEECH_CLIENT_FAILED_INIT_LOGGER = SPEECH_CLIENT_ERROR_BASE + 4
const SPEECH_CLIENT_RECV_API_ERR = SPEECH_CLIENT_ERROR_BASE + 5

// entity errors 1101 - 1200
const ENTITY_KEY_NOT_FOUND_REF_STRING = SPEECH_CLIENT_ERROR_BASE + 101
const ENTITY_VALUE_NOT_FOUND_REF_STRING = SPEECH_CLIENT_ERROR_BASE + 102
const ENTITYSET_FAILED_READ_ENTITIES_FILE = SPEECH_CLIENT_ERROR_BASE + 103
const ENTITY_NOT_FOUND_IN_ENTITYSET = SPEECH_CLIENT_ERROR_BASE + 104
const ENTITY_FILE_BLANK = SPEECH_CLIENT_ERROR_BASE + 105

// hotword errors 1201 - 1300
const HOTWORD_FAILED_GET_SPEECH_STDIN = SPEECH_CLIENT_ERROR_BASE + 201
const HOTWORD_FAILED_PORTAUDIO_READ = SPEECH_CLIENT_ERROR_BASE + 202
const HOTWORD_FAILED_BUF_WRITE_FROM_PORTAUDIO = SPEECH_CLIENT_ERROR_BASE + 203

// IPC PIPE errors 1301 - 1400
const IPC_FAILED_TO_INIT_CLIENT_LIST = SPEECH_CLIENT_ERROR_BASE + 301
const IPC_FAILED_TO_CREATE_TCP_SERVER = SPEECH_CLIENT_ERROR_BASE + 302
const IPC_FAILED_TO_START_LISTENER = SPEECH_CLIENT_ERROR_BASE + 303
const IPC_FAILED_TO_CONVERT_INCOMING_DATA = SPEECH_CLIENT_ERROR_BASE + 304
const IPC_FAILED_TO_INITIALIZE_SERVER = SPEECH_CLIENT_ERROR_BASE + 305
const IPC_FAILED_TO_CREATE_DATA_LIST = SPEECH_CLIENT_ERROR_BASE + 306

// hombot error errors 1401 - 1500
const HOMBOT_FAILED_TO_INIT_IPCCLIENT = SPEECH_CLIENT_ERROR_BASE + 401
const HOMBOT_NOT_ENOUGH_ARGS = SPEECH_CLIENT_ERROR_BASE + 402

// logger error errors 1501 - 1600
const LOGGER_FAILED_CREATE_LOG_FILE = SPEECH_CLIENT_ERROR_BASE + 501

//mqtt error codes 1601 - 1700

// HomBot server code 1701 - 1750
const HBSRV_FAILED_INIT_LOGGER = SPEECH_CLIENT_ERROR_BASE + 701

// executor error codes 1751 - 1800
const EX_FAILED_TO_LOAD_PLUGIN = SPEECH_CLIENT_ERROR_BASE + 751
const EX_FAILED_TO_LOOKUP_INIT_FUNCTION = SPEECH_CLIENT_ERROR_BASE + 752
const EX_FAILED_TO_LOOKUP_EXECUTE_FUNCTION = SPEECH_CLIENT_ERROR_BASE + 753
const EX_FAILED_TO_LOOKUP_DESTROY_FUNCTION = SPEECH_CLIENT_ERROR_BASE + 754
const EX_PLUGIN_NOT_FOUND = SPEECH_CLIENT_ERROR_BASE + 755

// Scheduler error codes 1801 - 1850
const TIMER_INVALID_DURATION = SPEECH_CLIENT_ERROR_BASE + 801

// ====================================================================
// ===============EXTERNAL ERROR CODES START FROM 5000=================
// ====================================================================
const EXTERNAL_COMP_BASE_ERR = 5000

// MQTT executor error codes 5000 - 5050
const MQTT_CONNECT_TIMEOUT = EXTERNAL_COMP_BASE_ERR + 1
const MQTT_CONNECT_ERROR = EXTERNAL_COMP_BASE_ERR + 2
const MQTT_ENTITY_NOT_FOUND_IN_SET = EXTERNAL_COMP_BASE_ERR + 3
const MQTT_VALUE_ID_NOT_FOUND = EXTERNAL_COMP_BASE_ERR + 4
const MQTT_SOURCE_ID_NOT_FOUND = EXTERNAL_COMP_BASE_ERR + 5
