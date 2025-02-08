package speech

import (
	"bytes"
	"sync"
)

type MicAudioStream struct {
	buffer bytes.Buffer
	//closing chan struct{}
	mutex sync.Mutex
}

func NewMicAudioStream() *MicAudioStream {
	return &MicAudioStream{
		//buffer:  make([][]int16, 0),
		//closing: make(chan struct{}),
	}
}

func (m *MicAudioStream) Write(srcBuf bytes.Buffer) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.buffer.Write(srcBuf.Bytes())
}

func (m *MicAudioStream) Read(data []byte) ([]byte, int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	//dataBuf := make([]byte, m.buffer.Len())
	nBytes, err := m.buffer.Read(dataBuf)
	//nBytes := copy(dataBuf, m.buffer.Read)
	//nBytes, err := io.Copy(bytes.NewBuffer(dataBuf), &m.buffer)
	// if err != nil {
	// 	return nil, -1, err
	// }
	return dataBuf, nBytes, err
}

func (m *MicAudioStream) Close() {
	//close(m.closing)
}

func (m *MicAudioStream) Clean() {
	m.buffer.Reset()
}
