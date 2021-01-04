package main

import "C"
import (
	"fmt"
	"time"

	"github.com/lonng/nano/connector"
	"github.com/lonng/nano/message"
	"github.com/lonng/nano/serialize/rawstring"
)

// RPC is to to request a rpc to server
//export RPC
func RPC(addrCChar *C.char, protocolNameCChar *C.char,
	requestCChar *C.char) (responseCChar *C.char) {

	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				response := fmt.Sprintf("{\"error\": \"%s\"}", e.Error())
				responseCChar = C.CString(response)
			}
		}
	}()

	addr := C.GoString(addrCChar)
	protocolName := C.GoString(protocolNameCChar)
	request := C.GoString(requestCChar)

	chData := make(chan string)
	conn := connector.NewConnector(
		connector.WithSerializer(rawstring.NewSerializer()),
	)
	defer conn.Close()
	conn.OnConnected(func(interface{}) {
		if err := conn.Request(protocolName, &request, func(data interface{}) {
			msg, ok := data.(*message.Message)
			if !ok {
				panic(fmt.Errorf("response is not *message.Message type"))
			}
			chData <- string(msg.Data)
		}); err != nil {
			panic(err)
		}
	})
	if err := conn.Start(addr); err != nil {
		panic(err)
	}

	var response string
	select {
	case <-time.After(time.Duration(10) * time.Second):
		panic(fmt.Errorf("request over time"))
	case response = <-chData:
		break
	}
	responseCChar = C.CString(response)
	return responseCChar
}

func main() {}
