package handlers

import (
	"encoding/json"
	"log"
	"net"
	"worker/internal/models"
	tcpMsgParser "worker/internal/protocol"
)

func DeleteTaskFromGlobalMap() {

}

func Receive(conn net.Conn, requestChan chan<- models.SearchRequest) { // отправлять таски в поток горутинам
	defer conn.Close()
	defer close(requestChan)
	for {
		data := make([]byte, 2028)
		if _, err := conn.Read(data); err != nil {
			log.Println("Error reading from connection:", err)
		}
		log.Println("Received data:", string(data))
		msg, err := tcpMsgParser.DecodePacket(data)
		if err != nil {
			log.Println("Error decoding packet:", err)
			continue
		}

		request := models.SearchRequest{}

		if err := json.Unmarshal(msg.Data_json, &request); err != nil {
			log.Println("Error unmarshalling JSON:", err)
			continue
		}

		requestChan <- request
	}
}

func Send(conn net.Conn, responseChan <-chan any) { //сюда напрямую пишут горутины
	defer conn.Close()
	for response := range responseChan {

		buf, err := tcpMsgParser.EncodePacket(0, 1, response)
		if err != nil {
			log.Println("Error encoding packet:", err)
			continue
		}

		if _, err := conn.Write(buf.Bytes()); err != nil {
			log.Println("Error writing to connection:", err)
			continue
		}
	}
}
