/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package coap

import (
	"bufio"
	"fmt"
	"net"
)

type CoapServer struct {
	l        net.Listener
	handlers map[string]Handler
	csm      *Capabilities
}

type Handler interface {
	Serve(peerIP net.Addr, packet *CoapPacket) *CoapPacket
}

func NewCoapServer() CoapServer {
	return NewCoapServerWithCSM(&Capabilities{10000, false})
}

func NewCoapServerWithCSM(csm *Capabilities) CoapServer {
	return CoapServer{handlers: map[string]Handler{}, csm: csm}
}

func (server *CoapServer) Start(address string, c chan bool) error {
	l, err := net.Listen("tcp4", address)
	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- false
		}
		return err
	}
	server.l = l

	fmt.Printf("CoapServer listening on %v\n", l.Addr())

	if c != nil {
		c <- true
	}
	defer server.l.Close()
	for {
		c, err := server.l.Accept()
		if err != nil {
			fmt.Println(err)
			return err
		} else {
			go server.handleConnection(c)
		}
	}

	return nil
}

func (server CoapServer) Stop() error {
	return server.l.Close()
}

func (server *CoapServer) Handle(uriPath string, handler Handler) {
	server.handlers[uriPath] = handler
}

func (server *CoapServer) HandleFunc(uriPath string, handler func(request *CoapPacket) *CoapPacket) {
	server.handlers[uriPath] = HandlerFunc(handler)
}

type HandlerFunc func(request *CoapPacket) *CoapPacket

func (f HandlerFunc) Serve(peerIP net.Addr, packet *CoapPacket) *CoapPacket {
	return f(packet)
}

func (server *CoapServer) handleConnection(c net.Conn) {
	fmt.Printf("Connected %v\n", c.RemoteAddr())
	reader := bufio.NewReader(c)
	//clientCapabilities := Capabilities{1152, false}

	//send server capabilities
	coapCSM := NewCoapPacket(CODE_701_CSM, []byte{}, []byte{})
	coapCSM.CSM = server.csm

	err := coapCSM.Write(c)
	if err != nil {
		fmt.Printf("Disconecting %v - %s\n", c.RemoteAddr(), err)
		c.Close()
		return
	}

	//wait for client CSM
	clientCSM, err := ReadCoap(reader)
	if err != nil {
		fmt.Printf("Disconecting %v - %s\n", c.RemoteAddr(), err)
		c.Close()
		return
	}

	fmt.Printf("CoAP server received: %v\n", clientCSM)
	if clientCSM.Code != CODE_701_CSM || clientCSM.CSM == nil {
		c.Close()
		return
	}

	for {
		req, err := ReadCoap(reader)
		if err != nil {
			fmt.Printf("Disconecting %v: %s\n", c.RemoteAddr(), err)
			c.Close()
			return
		}

		fmt.Printf("CoAP server received: %v\n", req)

		//ping
		if req.Code == CODE_702_PING {
			NewCoapPacket(CODE_703_PONG, []byte{}, []byte{}).Write(c)
			continue
		}

		//request
		if req.Code > 0 && req.Code <= 4 {
			handler, exists := server.handlers[req.UriPath]

			var resp *CoapPacket
			if exists {
				resp = handler.Serve(c.RemoteAddr(), req)
			} else {
				resp = NewCoapPacket(CODE_404_NOT_FOUND, req.Token, []byte{})
			}

			resp.Write(c)
			continue
		}

	}
	c.Close()
}
