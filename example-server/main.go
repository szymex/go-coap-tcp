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
	"github.com/szymex/go-coap-tcp/coap"
	"net"
	"time"
)

func main() {
	server := coap.NewCoapServer()
	server.HandleFunc("/time", func(req *coap.CoapPacket) *coap.CoapPacket {
		t := time.Now().In(time.UTC)
		return coap.NewCoapPacket(coap.CODE_205, req.Token, []byte(t.Format("2006-01-02 15:04:05 -0700 MST")))
	})

	server.Handle("/my-ip", MyIpHandler{})

	server.HandleFunc("/rfc8323", func(req *coap.CoapPacket) *coap.CoapPacket {
		return coap.NewCoapPacket(coap.CODE_205, req.Token, []byte(rfc8323))
	})

	panic(server.Start(":5683", nil))
}

type MyIpHandler struct {
}

func (f MyIpHandler) Serve(addr net.Addr, req *coap.CoapPacket) *coap.CoapPacket {
	ipAdr := addr.(*net.TCPAddr).IP

	return coap.NewCoapPacket(coap.CODE_205, req.Token, []byte(ipAdr.String()))
}
