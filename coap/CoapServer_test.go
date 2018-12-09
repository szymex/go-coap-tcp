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
	"testing"
)

func Test_ping_pong(t *testing.T) {

	server := NewCoapServer()
	start(&server, ":5683")

	client := connectClient(t, "127.0.0.1:5683")

	err := client.Ping()
	if err != nil {
		t.Fatal(err)
	}

	client.Close()
	server.Stop()
}

func Test_request_not_found(t *testing.T) {

	server := NewCoapServer()
	start(&server, ":15683")

	client := connectClient(t, "127.0.0.1:15683")

	resp, err := client.Get("/test")
	if err != nil {
		t.Fatal(err)
	}

	if resp.Code != CODE_404_NOT_FOUND {
		t.Fatalf("\nExpected: %#v \n  Actual: %#v", CODE_404_NOT_FOUND, resp)
	}

	client.Close()
	server.Stop()
}

func Test_request(t *testing.T) {

	server := NewCoapServer()
	server.HandleFunc("/test", func(req *CoapPacket) *CoapPacket {
		return NewCoapPacket(CODE_205_CONTENT, req.Token, []byte("test test"))
	})

	start(&server, ":25683")

	client := connectClient(t, "127.0.0.1:25683")

	resp, err := client.Get("/test")
	if err != nil {
		t.Fatal(err)
	}

	if resp.Code != CODE_205_CONTENT {
		t.Fatalf("\nExpected: 2.05\n  Actual: %v", resp.StringCode())
	}

	client.Close()
	server.Stop()
}

func Test_getHandlerShouldReturn405OnPost(t *testing.T) {

	server := NewCoapServer()
	server.HandleGet("/test", func(req *CoapPacket) *CoapPacket {
		return NewCoapPacket(CODE_205_CONTENT, req.Token, []byte("test test"))
	})

	start(&server, ":5683")
	client := connectClient(t, "127.0.0.1:5683")

	resp, _ := client.Post("/test", "")
	if resp.Code != CODE_405_METHOD_NOT_ALLOWED {
		t.Fatalf("\nExpected: 4.05\n  Actual: %v", resp.StringCode())
	}

	client.Close()
	server.Stop()
}

func start(server *CoapServer, address string) {
	ch := make(chan bool)
	go server.Start(address, ch)
	<-ch
}

func connectClient(t *testing.T, address string) *CoapClient {
	client, err := Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
