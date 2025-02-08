//#include <SoftwareSerial.h>
#include <EEPROM.h>
#include <ESP8266WiFi.h> // WiFi Library
#include <PubSubClient.h> // MQTT library

#include <ESP8266httpUpdate.h>

//static const uint8_t  PIN_D0  =  16;
static const uint8_t  PIN_D1  =  5;
static const uint8_t  PIN_D2  =  4;
static const uint8_t  PIN_D3  =  0;
static const uint8_t  PIN_D5  =  14;
static const uint8_t  PIN_D6  =  12;
static const uint8_t  PIN_D7  =  13;
static const uint8_t  PIN_D8  =  15;
static const uint8_t  PIN_D9  =  3;
//static const uint8_t  PIN_D10 =  1;
static const uint8_t  PIN_SD3 =  10;

//static const uint8_t  D0_MASK   = 1;
static const uint8_t  D1_MASK   = 2;
static const uint8_t  D2_MASK   = 4;
static const uint8_t  D3_MASK   = 8;
static const uint8_t  D5_MASK   = 16;
static const uint8_t  D6_MASK   = 32;
static const uint8_t  D7_MASK   = 64;
static const uint8_t  D8_MASK   = 128;
static const uint16_t  D9_MASK   = 256;
//static const uint16_t  D10_MASK  = 512;
static const uint16_t  SD3_MASK  = 1024;

static const char* pFirmware = "{\"fdate\":\"11-Aug-2019\",\"ftime\":\"11:45:00\",\"fver\":\"HBv1.0-11082019AM\"}";
//static const byte PSC_PAYLOAD = 1;

#define SUCCESS 0
#define FAILED -1

const char *ssid = "";   // cannot be longer than 32 characters!
const char *pass = ""; //

const int mqtt_port = 1883;
const char *mqtt_server = "192.168.29.204"; //Replace this with IP address of your Raspberry Pi where Mosquitto Broker is running
const char *mqtt_user = NULL;
const char *mqtt_pass = NULL;

const char *mqtt_client_name = "d-hombot-controller-con001";

const char *sub_topic = "hombot/type/controller/id/con001/evt/+/fmt/bin";
const char *monitor_topic = "hombot/type/controller/id/con001/monitor/msg/fmt/txt";
const char *error_topic = "hombot/type/controller/id/con001/err/msg/fmt/txt";
const char *monitor_state_topic = "hombot/type/controller/id/con001/monitor/state/fmt/bin";
//const char *default_topic = "hombot/type/controller/id/con001/evt/defaultState/fmt/bin";
//const char *psc_topic = "hombot/type/controller/id/con001/evt/psc/fmt/bin";
//const char *upate_topic = "hombot/type/controller/id/con001/evt/update/fmt/bin";

const char *psc_event = "psc";
const char *update_event = "update";
const char *default_event = "defaultState";
const char *firmware = "firmware";

#define PSC_EVT_LENGTH 3
#define UPDATE_EVT_LENGTH 6
#define DEFAULT_EVT_LENGTH 12
#define FIRMWARE_LENGTH 8

#define DEFUALT_STATE_ADDRESS 0
#define PUBMSG_BUFFER_SIZE 256

#define ALL_PINS_OFF  100532224

byte evt_pos = 0;

char pubStr[PUBMSG_BUFFER_SIZE];
int iLenTopic = 0;
char *pTopic = NULL;
unsigned int iPayLen = 0;
byte *pPayload = NULL;
 
WiFiClient wclient; //Declares a WifiClient Object using ESP8266WiFi
PubSubClient mqttclient(wclient); //instanciates client object

typedef struct UpdatePayload {
  const char *pAdd;
  short addLen;
  short port;
  const char *pEP;
  short epLen;
};
  

/* Callback to receive IoT commands
 *  
 *  Supported commands
 *    PSC -> pin state change
 *    Structure:
 *      topic: hombot/type/controller/id/con001/evt/psc/fmt/bin
 *    Payload:
 *      32 bit
 *      lower 2 bytes for PIN ON states
 *      higher 2 bytes for PIN OFF states
 */
void callback(const char* intopic, byte* inpayload, unsigned int payloadlength)
{
  Serial.printf("\nMessage arrived: %s, length=%d\n", intopic, payloadlength);
  //allocate for topic
  int iTpLen = strlen(intopic);
  if(iTpLen > iLenTopic) {
    iLenTopic = iTpLen;
    pTopic = (char*)realloc(pTopic, iLenTopic+1);
    if(NULL == pTopic)
      return;
  }
  strcpy(pTopic, intopic);

  //allocate for payload
  if(payloadlength > iPayLen) {
    iPayLen = payloadlength;
    pPayload = (byte*)realloc(pPayload, iPayLen+1);
    if(NULL == pPayload)
      return;
  }
  memcpy(pPayload, inpayload, payloadlength);
  pPayload[payloadlength] = 0;

  //*****************************************************************************
  //INTOPIC, INPAYLOAD pointers MUST NOT BE USED BEYOND THIS POINT
  //*****************************************************************************
  
  mqttclient.publish(monitor_topic, pTopic);

  if(strncmp(pTopic+evt_pos, psc_event, PSC_EVT_LENGTH) == 0)
  {
    //we have command PSC and
    //we are expecting a 32 bit UINT in BIG ENDIAN 
    if(payloadlength == 4) 
    {      
      uint32_t states = u32BEToLE(pPayload);      
      SetPinStates(states);
      states = GetPinStates();
      uint8_t arr[4];
      u32LEToByteArrayBE(&states, arr);
      //snprintf(pubStr, PUBMSG_BUFFER_SIZE, "State set to %d: hex %x,%x,%x,%x",states, arr[3],arr[2],arr[1],arr[0]);
      //mqttclient.publish(monitor_topic, pubStr);
      mqttclient.publish(monitor_state_topic, (const uint8_t*)arr, sizeof(arr));

    }
  }
  else if(strncmp(pTopic+evt_pos, default_event, DEFAULT_EVT_LENGTH) == 0)
  {
    Serial.printf("Setting defaultState");
    //we are expecting a 32 bit UINT in BIG ENDIAN 
    if(payloadlength == 4) 
    {      
      Serial.printf("Saving Default Payload %d,%d,%d,%d\n",pPayload[3],pPayload[2],pPayload[1],pPayload[0]);      
      EEPROM.write(DEFUALT_STATE_ADDRESS+0,pPayload[0]);
      EEPROM.write(DEFUALT_STATE_ADDRESS+1,pPayload[1]);
      EEPROM.write(DEFUALT_STATE_ADDRESS+2,pPayload[2]);
      EEPROM.write(DEFUALT_STATE_ADDRESS+3,pPayload[3]);
      EEPROM.commit();
      Serial.printf("Payload saved\n");

      snprintf(pubStr, PUBMSG_BUFFER_SIZE, "Default State set to: %d",  readDefaultState());
      mqttclient.publish(monitor_topic, pubStr);
    }    
  }
  else if(strncmp(pTopic+evt_pos, update_event, UPDATE_EVT_LENGTH) == 0)
  {
    Serial.printf("\nupdate requested %s len:%d", (const char*)pPayload, payloadlength);

    UpdatePayload *upPayload = (UpdatePayload*)malloc(sizeof(UpdatePayload));
    parseUpdatePayload(pPayload, payloadlength, upPayload);
    
    Serial.printf("\nCalling Update\n");
    t_httpUpdate_return ret = ESPhttpUpdate.update(upPayload->pAdd, upPayload->port, upPayload->pEP);
    Serial.printf("\nUpdate Called\n");
    switch(ret)
    {
      case HTTP_UPDATE_FAILED:
        Serial.printf("HTTP_UPDATE_FAILED %d, %s\n",ESPhttpUpdate.getLastError(), ESPhttpUpdate.getLastErrorString().c_str());
        snprintf(pubStr, PUBMSG_BUFFER_SIZE, "HTTP_UPDATE_FAILED: %d, %s",  
                                              ESPhttpUpdate.getLastError(),
                                              ESPhttpUpdate.getLastErrorString().c_str());
        mqttclient.publish(error_topic, pubStr);
        break;
      case HTTP_UPDATE_NO_UPDATES:
        Serial.printf("HTTP_UPDATE_NO_UPDATES\n");
        mqttclient.publish(error_topic, "HTTP_UPDATE_NO_UPDATES");
        break;
      case HTTP_UPDATE_OK:
        Serial.printf("HTTP_UPDATE_OK\n");
        mqttclient.publish(error_topic, "HTTP_UPDATE_OK");
        break;
    }
      
  } 
  else if(strncmp(pTopic+evt_pos, firmware, FIRMWARE_LENGTH) == 0)
  {
    snprintf(pubStr, PUBMSG_BUFFER_SIZE, 
                      "Firmware: %s", 
                      "fdate:11-Aug-2019,ftime:13:30:00,fver:HBv1.0-11082019-1330");
                      
                      //"{\"fdate\":\"11-Aug-2019\",\"ftime\":\"11:45:00\",\"fver\":\"HBv1.0-11082019AM\"}");
      
    mqttclient.publish(monitor_topic, pubStr);  
  }
}

boolean parseUpdatePayload(byte *updatePayload, 
                           unsigned int len,
                           UpdatePayload *upPayload) 
{
  //Serial.printf("\nval = %d,%d,%d,%d,%d,%d",*(updatePayload+0),*(updatePayload+1),*(updatePayload+2),*(updatePayload+3),*(updatePayload+4),*(updatePayload+5));
  short pos=0;
  short len1 = getShortLE(updatePayload+pos);
  upPayload->pAdd = (const char*)(updatePayload+2);
  upPayload->addLen = len1;
  //Serial.printf("\naddlen %d\n", addlen);
  //Serial.printf("\nadd %s", upPayload->pAdd);

  pos = pos+2+len1+2;
  upPayload->port = getShortLE(updatePayload+pos);
  //Serial.printf("\nPort %d", upPayload->port);  

  pos = pos+2;
  len1 = getShortLE(updatePayload+pos);
  upPayload->pEP = (const char*)(updatePayload+pos+2);
  upPayload->epLen = len1;
  
 //Serial.printf("\npayload Add=%s Len=%d, Port=%d, EP=%s Len=%d", upPayload->pAdd, upPayload->addLen, upPayload->port, upPayload->pEP, upPayload->epLen);
  return true;
}

uint32_t u32BEToLE(byte* payload) {
      uint32_t states = 0;
      states += payload[0] << 24;
      states += payload[1] << 16;
      states += payload[2] << 8;
      states += payload[3];
      Serial.printf("State value is %d\n",states);  
      return states;
}

uint32_t u32LEToBE(uint32_t* payload) {
      uint32_t states = 0;
      states += *((uint8_t*)payload + 0) << 24;
      states += *((uint8_t*)payload + 1) << 16;
      states += *((uint8_t*)payload + 2) << 8;
      states += *((uint8_t*)payload + 3);
      return states;
}

uint8_t* u32LEToByteArrayBE(uint32_t* payload, uint8_t* pArr) {
      *(pArr + 3) = *((uint8_t*)payload + 0);
      *(pArr + 2) = *((uint8_t*)payload + 1);
      *(pArr + 1) = *((uint8_t*)payload + 2);
      *(pArr + 0) = *((uint8_t*)payload + 3);
      return pArr;
}

short getShortLE(byte* payload) {
  uint16_t val = 0;
  val += *(payload + 0) << 8;
  val += *(payload + 1);

  return val;
}

uint32_t readDefaultState() {
    byte* payload = (byte*)malloc(4);
    payload[0] = EEPROM.read(DEFUALT_STATE_ADDRESS+0);
    payload[1] = EEPROM.read(DEFUALT_STATE_ADDRESS+1);
    payload[2] = EEPROM.read(DEFUALT_STATE_ADDRESS+2);
    payload[3] = EEPROM.read(DEFUALT_STATE_ADDRESS+3);
    Serial.printf("Getting states from saved payload %d,%d,%d,%d\n",payload[3],payload[2],payload[1],payload[0]);
    uint32_t states = u32BEToLE(payload);
    free(payload);  
    return states;
}

void setup() {
  EEPROM.begin(64);

  //configure for telnet
//  Serial.swap();
//  // Hardware serial is now on RX:GPIO13 TX:GPIO15
//  // use SoftwareSerial on regular RX(3)/TX(1) for logging
//  logger = new SoftwareSerial(3, 1);
//  logger->begin(BAUD_LOGGER);
//  logger->println("\n\nUsing SoftwareSerial for logging");  
  

  int mqttstate = 0;
  // Setup console
  Serial.begin(115200,SERIAL_8N1,SERIAL_TX_ONLY);  //set the baud rate
  delay(10);

  Serial.printf("Booting the updated firmware: %s\n", pFirmware);

  const char *pos = strchr(sub_topic,'+');
  evt_pos = (pos - sub_topic);
  
  mqttclient.setServer(mqtt_server, mqtt_port);
  mqttclient.setCallback(callback);
  
  //The first thing we do is set all the PINS mode
  //We found issues with PIN_D0 and D10, we will set them to LOW directly
  //pinMode(PIN_D0, OUTPUT);
  //pinMode(PIN_D10, OUTPUT);
  SetPinStates(ALL_PINS_OFF);
  
  pinMode(PIN_D1, OUTPUT);
  pinMode(PIN_D2, OUTPUT);
  pinMode(PIN_D3, OUTPUT);
  pinMode(PIN_D5, OUTPUT);
  pinMode(PIN_D6, OUTPUT);
  pinMode(PIN_D7, OUTPUT);
  pinMode(PIN_D8, OUTPUT);
  pinMode(PIN_D9, OUTPUT);
  pinMode(PIN_SD3, OUTPUT);


  uint32_t states = readDefaultState();
  SetPinStates(states);


//  byte payloadType = EEPROM.read(0);  
//  if(PSC_PAYLOAD == payloadType)
//  {
//    //we have command PSC and
//    //we are expecting a 32 bit UINT in BIG ENDIAN 
//  }
  
  //do first wifi connect here
  if(WiFi.getAutoConnect() == false)
    WiFi.setAutoReconnect(true);

  connectWifi();

  if (WiFi.status() == WL_CONNECTED) 
  {
    if((mqttstate = connectMqtt()) == SUCCESS)
      Serial.println("Connected successfully to mqtt server ");
    else
    {
      Serial.print("\nFailed to connect to mqtt server:");
      printMqttClientState(mqttstate);
    }
      
  }  
}

void loop() {
  if (WiFi.status() == WL_CONNECTED) 
  {
    if(connectMqtt() == SUCCESS) 
    {
      mqttclient.loop(); 
    }
  }
  else
  {
    connectWifi();
  }
  delay(200);

}

//====================================================================
//====================================================================
void connectWifi() 
{
    //wifi not connected?
    Serial.print("Re-Connecting to ");
    Serial.printf("%s...\n", ssid);
    
    WiFi.begin(ssid, pass);
    int status = WiFi.waitForConnectResult();
    if (status  != WL_CONNECTED) {
      Serial.print("WiFi state is ");
      printWifiState(status);
      Serial.println(WiFi.macAddress());
      return;
    }
    else    
    {
      Serial.print("WiFi connected. IP = ");
      Serial.println(WiFi.localIP());
    }  
}

void printWifiState(int status) {
  switch(status) {
    case WL_NO_SSID_AVAIL:
      Serial.println("WL_NO_SSID_AVAIL");
      break;
    case WL_CONNECT_FAILED:
      Serial.println("WL_CONNECT_FAILED");
      break;
    case WL_IDLE_STATUS:
      Serial.println("WL_IDLE_STATUS");
      break;
    case WL_DISCONNECTED:
      Serial.println("WL_DISCONNECTED");
      break;
    default:
      Serial.println("Unknown state");
      break;
  }
}

void printMqttClientState(int istate) 
{
  Serial.print("MQTT State: ");
  switch(istate) 
  {
  //    0 : MQTT_CONNECTED - the client is connected
  case MQTT_CONNECTED:
    Serial.println("MQTT_CONNECTED");
    break;    
  //    -4 : MQTT_CONNECTION_TIMEOUT - the server didn't respond within the keepalive time
  case MQTT_CONNECTION_TIMEOUT:
    Serial.println("MQTT_CONNECTION_TIMEOUT");
    break;
  //    -3 : MQTT_CONNECTION_LOST - the network connection was broken
  case MQTT_CONNECTION_LOST:
    Serial.println("MQTT_CONNECTION_LOST");
    break;
  //    -2 : MQTT_CONNECT_FAILED - the network connection failed
  case MQTT_CONNECT_FAILED:
    Serial.println("MQTT_CONNECT_FAILED");
    break;
  //    -1 : MQTT_DISCONNECTED - the client is disconnected cleanly
  case MQTT_DISCONNECTED:
    Serial.println("MQTT_DISCONNECTED");
    break;
  //    1 : MQTT_CONNECT_BAD_PROTOCOL - the server doesn't support the requested version of MQTT
  case MQTT_CONNECT_BAD_PROTOCOL:
    Serial.println("MQTT_CONNECT_BAD_PROTOCOL");
    break;
  //    2 : MQTT_CONNECT_BAD_CLIENT_ID - the server rejected the client identifier
  case MQTT_CONNECT_BAD_CLIENT_ID:
    Serial.println("MQTT_CONNECT_BAD_CLIENT_ID");
    break;
  //    3 : MQTT_CONNECT_UNAVAILABLE - the server was unable to accept the connection
  case MQTT_CONNECT_UNAVAILABLE:
    Serial.println("MQTT_CONNECT_UNAVAILABLE");
    break;
  //    4 : MQTT_CONNECT_BAD_CREDENTIALS - the username/password were rejected
  case MQTT_CONNECT_BAD_CREDENTIALS:
    Serial.println("MQTT_CONNECT_BAD_CREDENTIALS");
    break;
  //    5 : MQTT_CONNECT_UNAUTHORIZED - the client was not authorized to connect
  case MQTT_CONNECT_UNAUTHORIZED:
    Serial.println("MQTT_CONNECT_UNAUTHORIZED");
    break;
  default:
    Serial.println("MQTT_UNKNOWN");
    break;
  }
}

int connectMqtt() 
{
    //client object makes connection to server
    if (!mqttclient.connected()) 
    {
      Serial.printf("\nConnecting to MQTT server %s:%d as %s\n", 
                                        mqtt_server, 
                                        mqtt_port,
                                        mqtt_client_name);
      //Authenticating the client object
      if (mqttclient.connect(mqtt_client_name,
                             mqtt_user,
                             mqtt_pass)) 
      {
        Serial.println("Connected to MQTT server");
        //Subscribe code
        if(mqttclient.subscribe(sub_topic) != true) 
        {
          Serial.printf("\nSubscribe to topic %s Failed\n", sub_topic);
          int istate = mqttclient.state();
          return istate;
        }
      }
      else 
      {
        int istate = mqttclient.state();
        Serial.println("Connection to MQTT server failed with ");
        printMqttClientState(istate);
        return istate;
      }
    }
    
    return SUCCESS;
}

//=====================================================================
// SECTION HANDLING PIN STATES
//=====================================================================
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
 *  
 */

uint32_t GetPinStates() 
{
  uint32_t states = 0;
  
//  if(digitalRead(PIN_D0) == HIGH)
//    states = states | D0_MASK;

  if(digitalRead(PIN_D1) == LOW)
    states = states | D1_MASK;
  else
    states = states | (D1_MASK << 16);
    
  if(digitalRead(PIN_D2) == LOW)
    states = states | D2_MASK;
  else
    states = states | (D2_MASK << 16);

  if(digitalRead(PIN_D3) == LOW)
    states = states | D3_MASK;
  else
    states = states | (D3_MASK << 16);

  if(digitalRead(PIN_D5) == LOW)
    states = states | D5_MASK;
  else
    states = states | (D5_MASK << 16);

  if(digitalRead(PIN_D6) == LOW)
    states = states | D6_MASK;
  else
    states = states | (D6_MASK << 16);

  if(digitalRead(PIN_D7) == LOW)
    states = states | D7_MASK;
  else
    states = states | (D7_MASK << 16);
    
  if(digitalRead(PIN_D8) == LOW)
    states = states | D8_MASK;
  else
    states = states | (D8_MASK << 16);

  if(digitalRead(PIN_D9) == LOW)
    states = states | D9_MASK;
  else
    states = states | (D9_MASK << 16);

//  if(digitalRead(PIN_D10) == HIGH)
//    states = states | D10_MASK;

  if(digitalRead(PIN_SD3) == LOW)
    states = states | SD3_MASK;
  else
    states = states | (SD3_MASK << 16);

    return states;
}

void SetPinStates(uint32_t pinStates) 
{
  //mask off the higher byte
  uint16_t highByteMask = 0xFFFF;
  
  uint16_t offStates = 0;
  uint16_t onStates = 0;

  //get the lower 2 bytes
  onStates = pinStates & highByteMask;
  offStates = pinStates >> 16;

  Serial.printf("offstates=%d\t", offStates);
  Serial.printf("onstates=%d\n", onStates);

  SetPinStateOff(offStates);
  delay(200);
  SetPinStateOn(onStates);

}

//Set the state of each pin to HIGH
void SetPinStateOff(uint16_t pinState) 
{
//    Serial.printf("On state is %d\n", pinState);
//    Serial.printf("maskD0 AND %d is %d\n", pinState, pinState  & D0_MASK);
//    Serial.printf("maskD1 AND %d is %d\n", pinState, pinState  & D1_MASK);
//    Serial.printf("maskD2 AND %d is %d\n", pinState, pinState  & D2_MASK);
//    Serial.printf("maskD3 AND %d is %d\n", pinState, pinState  & D3_MASK);
    //get the state of each pin
//    if((pinState & D0_MASK) > 0) 
//    {
//      digitalWrite(PIN_D0, HIGH);
//    }
    
    if((pinState & D1_MASK) > 0) 
    {
      digitalWrite(PIN_D1, HIGH);
    }
    
    if((pinState & D2_MASK) > 0) 
    {
      digitalWrite(PIN_D2, HIGH);
    }
    
    if((pinState & D3_MASK) > 0) 
    {
      digitalWrite(PIN_D3, HIGH);
    }
    
    if((pinState & D5_MASK) > 0) 
    {
      digitalWrite(PIN_D5, HIGH);
    }
    
    if((pinState & D6_MASK) > 0) 
    {
      digitalWrite(PIN_D6, HIGH);
    }
    
    if((pinState & D7_MASK) > 0) 
    {
      //Serial.printf("PIN_D7 is HIGH");
      digitalWrite(PIN_D7, HIGH);
    }

    if((pinState & D8_MASK) > 0) 
    {
      digitalWrite(PIN_D8, HIGH);
    }
    
    if((pinState & D9_MASK) > 0) 
    {
      digitalWrite(PIN_D9, HIGH);
    }
    
//    if((pinState & D10_MASK) > 0) 
//    {
//      digitalWrite(PIN_D10, HIGH);
//    }

    if((pinState & SD3_MASK) > 0) 
    {
      digitalWrite(PIN_SD3, HIGH);
    }
}

//Set the state of each pin to LOW
void SetPinStateOn(uint16_t pinState) 
{
    
//    if((pinState & D0_MASK) > 0) 
//    {
//      digitalWrite(PIN_D0, LOW);
//    }
    
    if((pinState & D1_MASK) > 0) 
    {
      digitalWrite(PIN_D1, LOW);
    }
    
    if((pinState & D2_MASK) > 0) 
    {
      digitalWrite(PIN_D2, LOW);
    }
    
    if((pinState & D3_MASK) > 0) 
    {
      digitalWrite(PIN_D3, LOW);
    }
    
    if((pinState & D5_MASK) > 0) 
    {
      digitalWrite(PIN_D5, LOW);
    }
    
    if((pinState & D6_MASK) > 0) 
    {
      digitalWrite(PIN_D6, LOW);
    }
    
    if((pinState & D7_MASK) > 0) 
    {
      //Serial.printf("PIN_D7 is LOW");
      digitalWrite(PIN_D7, LOW);
    }

    if((pinState & D8_MASK) > 0) 
    {
      digitalWrite(PIN_D8, LOW);
    }
    
    if((pinState & D9_MASK) > 0) 
    {
      digitalWrite(PIN_D9, LOW);
    }
    
//    if((pinState & D10_MASK) > 0) 
//    {
//      digitalWrite(PIN_D10, LOW);
//    }

    if((pinState & SD3_MASK) > 0) 
    {
      digitalWrite(PIN_SD3, LOW);
    }
}
