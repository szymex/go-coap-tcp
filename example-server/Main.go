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
package main

import (
	"fmt"
	"github.com/szymex/go-coap-tcp/coap"
	"math/rand"
	"net"
	"time"
)

func main() {
	server := coap.NewCoapServer()
	server.HandleGet("/time", func(req *coap.CoapPacket) *coap.CoapPacket {
		t := time.Now().In(time.UTC)
		return coap.NewCoapPacket(coap.CODE_205_CONTENT, req.Token, []byte(t.Format("2006-01-02 15:04:05 -0700 MST")))
	})

	server.Handle("/my-ip", MyIpHandler{})

	server.HandleFunc("/rfc8323", func(req *coap.CoapPacket) *coap.CoapPacket {
		return coap.NewCoapPacket(coap.CODE_205_CONTENT, req.Token, []byte(rfc8323))
	})

	server.Handle("/tmp", &ReadWriteResourceHandler{})

	server.HandleGet("/slow", func(req *coap.CoapPacket) *coap.CoapPacket {
		wait := time.Duration(rand.Intn(9)) + 1
		time.Sleep(wait * time.Second)
		return coap.NewCoapPacket(coap.CODE_205_CONTENT, req.Token, []byte(fmt.Sprintf("Waited %d seconds", wait)))
	})

	panic(server.Start(":5683", nil))
}

type MyIpHandler struct {
}

func (f MyIpHandler) Serve(addr net.Addr, req *coap.CoapPacket) *coap.CoapPacket {
	ipAdr := addr.(*net.TCPAddr).IP

	return coap.NewCoapPacket(coap.CODE_205_CONTENT, req.Token, []byte(ipAdr.String()))
}

type ReadWriteResourceHandler struct {
	payload       []byte
	contentFormat int16
	maxAge        uint32
}

func (f *ReadWriteResourceHandler) Serve(addr net.Addr, req *coap.CoapPacket) *coap.CoapPacket {
	resp := coap.NewCoapPacket(coap.CODE_205_CONTENT, req.Token, []byte{})

	switch req.Code {
	case coap.GET:
		resp.Payload = f.payload
		resp.ContentFormat = f.contentFormat
		resp.MaxAge = f.maxAge

	case coap.PUT | coap.POST:
		f.maxAge = req.MaxAge
		f.contentFormat = req.ContentFormat
		f.payload = req.Payload
		resp.Code = coap.CODE_204_CHANGED

	case coap.DELETE:
		f.maxAge = 60
		f.contentFormat = -1
		f.payload = []byte{}
		resp.Code = coap.CODE_202_DELETED
	default:
		resp.Code = coap.CODE_500_INTERNAL_SERVER_ERROR
	}

	return resp
}
