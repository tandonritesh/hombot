package ipcpipe

import (
	"container/list"
	"hombot/errors"
	"hombot/utils/constants"
	"log"
	"net"
)

// package level variables
var dataRead chan []byte
var dataWrite chan []byte
var clientList *list.List

var CONVERT_ERROR []byte

var dataCallback func(buf []byte)

// ===================================================================
type Client struct {
	conn    net.Conn
	address net.Addr
}

func (c *Client) GetAddr() net.Addr {
	return c.address
}

func (c *Client) GetConn() net.Conn {
	return c.conn
}

// ===================================================================
type Data struct {
	address net.Addr
	bufSize int16
	buf     []byte
}

func (d *Data) GetAddr() net.Addr {
	return d.address
}

func (d *Data) GetBufSize() int16 {
	return d.bufSize
}

func (d *Data) GetBuffer() *[]byte {
	return &d.buf
}

// ===================================================================
func init() {
	dataRead = make(chan []byte)
	dataWrite = make(chan []byte)

	// errStr := fmt.Sprintf("%s%d", "Conversion Error-", errors.IPC_FAILED_TO_CONVERT_INCOMING_DATA)
	// CONVERT_ERROR = bytes.NewBufferString(errStr).Bytes()
}

func Init(mode string, cb func(buf []byte)) int {
	clientList = list.New()
	if clientList == nil {
		log.Fatalln("Failed to initialize clients list")
		return errors.IPC_FAILED_TO_INIT_CLIENT_LIST
	}

	dataCallback = cb

	//start the server in a go routine
	go start(mode)
	go ReadMessages()

	return errors.SUCCESS
}

func readData() []byte {
	return <-dataRead
}

func ReadMessages() {
	var dataBuf []byte
	for {
		dataBuf = readData()
		dataCallback(dataBuf[0:])
	}
}

func start(mode string) int {
	var conn net.Conn
	var err error
	var listner net.Listener
	var cl Client

	if mode == constants.SERVER {
		log.Printf("Starting Server at %s...", constants.PORT)
		listner, err = net.Listen("tcp4", constants.PORT)
		if err != nil {
			log.Fatalf("Error creating tcp server: %v", err)
			return errors.IPC_FAILED_TO_CREATE_TCP_SERVER
		}

		for {
			log.Println("Now Accepting Connection")
			conn, err = listner.Accept()
			if err != nil {
				log.Fatalf("Error starting listner: %v", err)
				return errors.IPC_FAILED_TO_START_LISTENER
			}
			cl.conn = conn
			cl.address = conn.RemoteAddr()
			clientList.PushBack(cl)

			log.Printf("Accepted Conn: Local: %v, Remote: %v", cl.conn.LocalAddr(), cl.conn.RemoteAddr())
			go readSocket(&cl)
			go writeSocket(&cl)
		}
	}
	return errors.SUCCESS
}

func readSocket(cl *Client) {
	var byteRead int
	var err error
	var buf []byte = make([]byte, 1024)

	log.Printf("Now reading for Local: %v, Remote: %v", cl.conn.LocalAddr(), cl.conn.RemoteAddr())
	for {
		byteRead, err = cl.conn.Read(buf)
		if err != nil {
			log.Printf("Error reading buffer: %v", err)
			cl.conn.Close()
			break
		}
		if byteRead > 0 {
			dataRead <- buf[:byteRead]
		}
	}
}

func writeSocket(cl *Client) {
	var buf []byte
	//var byteWrite int
	var err error

	buf = <-dataWrite
	_, err = cl.conn.Write(buf)
	if err != nil {
		log.Printf("Error writing data to socket: %v", err)
	}
	//log.Println("Data written to socket %d", byteWrite)
}
