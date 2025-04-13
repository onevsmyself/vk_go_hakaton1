package tcpMsgParser

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"

	"github.com/labstack/gommon/log"
)

const (
	newTask = 0
	tasks   = 1
	task    = 2
	delete  = 3
)

type Message struct {
	MsgID     uint32
	Seq       uint32
	Data_len  uint32
	Data_json []byte
	DataCRC   uint32
}

func EncodePacket(msgID uint32, seq uint32, payload any) (*bytes.Buffer, error) {
	buff := new(bytes.Buffer)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Cannot parse json, err: %s", err)
		return buff, err
	}

	sum := crc32.ChecksumIEEE(jsonData)

	binary.Write(buff, binary.BigEndian, msgID)
	binary.Write(buff, binary.BigEndian, seq)
	binary.Write(buff, binary.BigEndian, uint32(len(jsonData)))
	buff.Write(jsonData)
	binary.Write(buff, binary.BigEndian, sum)
	return buff, nil
}

func DecodePacket(data []byte) (Message, error) {
	var m Message
	buf := bytes.NewReader(data)

	binary.Read(buf, binary.BigEndian, &m.MsgID)
	binary.Read(buf, binary.BigEndian, &m.Seq)

	binary.Read(buf, binary.BigEndian, &m.Data_len)

	m.Data_json = make([]byte, m.Data_len)
	buf.Read(m.Data_json)

	binary.Read(buf, binary.BigEndian, &m.DataCRC)

	if crc32.ChecksumIEEE(m.Data_json) != m.DataCRC {
		return m, errors.New("CRC32 mismatch")
	}

	return m, nil
}
