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
	"errors"
	"fmt"
	"net"
)

func Connect(address string) (*CoapClient, error) {
	return ConnectWithCSM(address, &Capabilities{10000, false})
}

func ConnectWithCSM(address string, csm *Capabilities) (*CoapClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	client := CoapClient{conn, nil}

	//send capabilities
	coapCSM := NewCoapPacket(CODE_701_CSM, []byte{}, []byte{})
	coapCSM.CSM = csm

	err = coapCSM.Write(client.conn)
	fmt.Printf("    Sent: %v\n", coapCSM)
	if err != nil {
		conn.Close()
		return nil, err
	}

	//read capabilities
	peerCoap, errr := ReadCoap(client.conn)
	if errr != nil {
		conn.Close()
		return nil, errr
	}

	fmt.Printf("Received: %v\n", peerCoap)
	if peerCoap.Code != CODE_701_CSM || peerCoap.CSM == nil {
		conn.Close()
		return nil, errors.New("expecting csm not received")
	}

	client.serverCsm = peerCoap.CSM

	return &client, nil
}

func (client *CoapClient) Close() error {
	return client.conn.Close()
}

type CoapClient struct {
	conn      net.Conn
	serverCsm *Capabilities
}

func (client *CoapClient) Ping() error {
	coapPing := NewCoapPacket(CODE_702_PING, []byte{}, []byte{})

	err := coapPing.Write(client.conn)
	if err != nil {
		return err
	}

	resp, err := ReadCoap(client.conn)
	if err != nil {
		return err
	}

	if resp.Code != CODE_703_PONG {
		return errors.New("not expected response")
	}
	return nil
}

func (client *CoapClient) Get(uriPath string) (*CoapPacket, error) {
	return client.Invoke(GET, uriPath, -1, []byte{})
}

func (client *CoapClient) Post(uriPath string, payload string) (*CoapPacket, error) {
	return client.Invoke(POST, uriPath, MT_TEXT_PLAIN, []byte(payload))
}

func (client *CoapClient) Put(uriPath string, payload string) (*CoapPacket, error) {
	return client.Invoke(PUT, uriPath, MT_TEXT_PLAIN, []byte(payload))
}

func (client *CoapClient) Delete(uriPath string) (*CoapPacket, error) {
	return client.Invoke(DELETE, uriPath, -1, []byte{})
}

func (client *CoapClient) Invoke(method uint8, uriPath string, mediaType int16, payload []byte) (*CoapPacket, error) {
	req := NewCoapPacket(method, []byte{}, payload)
	req.UriPath = uriPath
	req.ContentFormat = mediaType

	err := req.Write(client.conn)
	if err != nil {
		return nil, err
	}
	fmt.Printf("    Sent: %v\n", req)

	//todo: verify token
	resp, err := ReadCoap(client.conn)
	if resp != nil {
		fmt.Printf("Received: %v\n", resp)
	}

	return resp, err
}
