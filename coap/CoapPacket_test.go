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
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestReadCoap_simplest(t *testing.T) {

	coap, _ := readCoap([]byte{0x00, 0x45})

	assert(t, NewCoapPacket(CODE_205_CONTENT, []byte{}, []byte{}), coap)
}

func TestReadCoap_withToken(t *testing.T) {

	coap, _ := readCoap([]byte{0x01, 0x43, 0x7f})

	assert(t, NewCoapPacket(CODE_203_VALID, []byte{0x7f}, []byte{}), coap)
}

func TestReadCoap_withPayload(t *testing.T) {

	coap, _ := readCoap([]byte{0x30, 0x41, 0xFF, 0x01, 0x02})

	assert(t, NewCoapPacket(CODE_201_CREATED, []byte{}, []byte{0x01, 0x02}), coap)

}

func TestReadCoap_withLargerPayload(t *testing.T) {

	coap, _ := readCoap([]byte{0xd0, 0x03, 0x41, 0xFF, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f})

	assert(t, NewCoapPacket(CODE_201_CREATED, []byte{}, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}), coap)

}

func TestCoapWithLargePayloads(t *testing.T) {

	coap := NewCoapPacket(CODE_412_PRECONDITION_FAILED, []byte{12}, make([]byte, 260))
	assert(t, coap, writeAndRead(coap, t))

	coap2 := NewCoapPacket(CODE_415_UNSUPPORTED_CONTENT_FORMAT, []byte{10, 20}, make([]byte, 65800))
	assert(t, coap2, writeAndRead(coap2, t))

	coap3 := NewCoapPacket(CODE_504_GATEWAY_TIMEOUT, []byte{10, 20, 30}, make([]byte, 200000))
	assert(t, coap3, writeAndRead(coap3, t))
}

func TestReadCoap_withUriPath(t *testing.T) {

	coap, _ := readCoap([]byte{0x70, 0x43, 0xb4, 't', 'e', 's', 't', 0x01, '2'})

	expectedPacket := NewCoapPacket(CODE_203_VALID, []byte{}, []byte{})
	expectedPacket.UriPath = "/test/2"
	assert(t, expectedPacket, coap)
}

func TestReadCoap_withMaxAge(t *testing.T) {

	coap, _ := readCoap([]byte{0x40, 0x43, 0xd2, 0x01, 0x01, 0xF0})

	expectedPacket := NewCoapPacket(CODE_203_VALID, []byte{}, []byte{})
	expectedPacket.MaxAge = 0x01F0
	assert(t, expectedPacket, coap)
}

func TestReadCoap_withContentFormat(t *testing.T) {

	coap, _ := readCoap([]byte{0x20, 0x43, 0xc1, 42})

	expectedPacket := NewCoapPacket(CODE_203_VALID, []byte{}, []byte{})
	expectedPacket.ContentFormat = MT_APPLICATION_OCTET_STREAM
	assert(t, expectedPacket, coap)
}

func TestWriteCoap_simplest(t *testing.T) {

	coap := NewCoapPacket(CODE_205_CONTENT, []byte{}, []byte{})
	w := new(bytes.Buffer)

	if coap.Write(w) != nil {
		t.Errorf("Error")
	}

	expected := []byte{0x00, 0x45}
	if !(bytes.Equal(w.Bytes(), expected)) {
		t.Errorf("Wrong: %#v", w.Bytes())
	}
}

func TestWriteCoap_token_and_payload(t *testing.T) {

	coap := NewCoapPacket(CODE_205_CONTENT, []byte{0x01, 0x02}, []byte{0x10, 0x11, 0x12})
	w := new(bytes.Buffer)

	if coap.Write(w) != nil {
		t.Errorf("Error")
	}

	expected := []byte{0x42, 0x45, 0x01, 0x02, 0xff, 0x10, 0x11, 0x12}
	if !(bytes.Equal(w.Bytes(), expected)) {
		t.Errorf("Wrong: %#v", w.Bytes())
	}
}

func TestReadWriteCoap(t *testing.T) {

	//simple
	coap := NewCoapPacket(CODE_205_CONTENT, []byte{}, []byte{})
	assert(t, coap, writeAndRead(coap, t))

	//token and payload
	coap2 := NewCoapPacket(CODE_205_CONTENT, []byte{0x01, 0x02}, []byte("test"))
	assert(t, coap2, writeAndRead(coap2, t))

	//all options
	coap2.ContentFormat = MT_APPLICATION_OCTET_STREAM
	coap2.MaxAge = 100
	coap2.UriPath = "/path1/long-path-long-path"
	assert(t, coap2, writeAndRead(coap2, t))

}

func TestCSM(t *testing.T) {

	coap := NewCoapPacket(CODE_701_CSM, []byte{}, []byte{})
	coap.CSM = &Capabilities{123, true}

	assert(t, coap, writeAndRead(coap, t))
}

func writeAndRead(coap *CoapPacket, t *testing.T) CoapPacket {
	w := new(bytes.Buffer)
	if coap.Write(w) != nil {
		t.Errorf("Error")
	}
	//fmt.Printf("coap: %x\n", w.Bytes())
	coap2, _ := readCoap(w.Bytes())
	return coap2
}

//------ helper test functions ----------
func readCoap(rawCoap []byte) (CoapPacket, error) {

	var reader io.Reader = bytes.NewBuffer(rawCoap)

	coap, err := ReadCoap(reader)
	return *coap, err
}

func assert(t *testing.T, expectedCoap *CoapPacket, actualCoap CoapPacket) {

	if !(expectedCoap.Code == actualCoap.Code &&
		bytes.Equal(expectedCoap.Token, actualCoap.Token) &&
		bytes.Equal(expectedCoap.Payload, actualCoap.Payload) &&
		expectedCoap.UriPath == actualCoap.UriPath &&
		expectedCoap.MaxAge == actualCoap.MaxAge &&
		reflect.DeepEqual(expectedCoap.CSM, actualCoap.CSM)) {

		t.Errorf("\nExpected: %v \n  Actual: %s", expectedCoap, actualCoap.String())
	}

}
