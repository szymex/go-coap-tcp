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
	"fmt"
	"io"
	"strings"
)

// https://tools.ietf.org/html/rfc8323#page-7

type CoapPacket struct {
	Code    uint8
	Token   []byte
	Payload []byte

	//options
	UriPath       string
	MaxAge        uint32
	ContentFormat int16

	CSM *Capabilities
}

type Capabilities struct {
	MaxMessageSize    uint32
	BlockWiseTransfer bool
}

func NewCoapPacket(code uint8, token []byte, payload []byte) *CoapPacket {
	return &CoapPacket{code, token, payload, "", 60, -1, nil}
}

const (
	GET    = 1
	POST   = 2
	PUT    = 3
	DELETE = 4

	c2xx = 2 << 5
	c4xx = 4 << 5
	c5xx = 5 << 5
	c7xx = 7 << 5

	CODE_201 = c2xx + 1
	CODE_202 = c2xx + 2
	CODE_203 = c2xx + 3 //0x43
	CODE_204 = c2xx + 4
	CODE_205 = c2xx + 5

	CODE_400_BAD_REQUEST = c4xx + 0
	CODE_404_NOT_FOUND   = c4xx + 4

	CODE_701_CSM  = c7xx + 1
	CODE_702_PING = c7xx + 2
	CODE_703_PONG = c7xx + 3

	CT_TEXT_PLAIN               = 0
	CT_APPLICATION_OCTET_STREAM = 42
)

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  Len  |  TKL  | Extended Length (if any, as chosen by Len) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |      Code     | Token (if any, TKL bytes) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  Options (if any) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |1 1 1 1 1 1 1 1|    Payload (if any) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func ReadCoap(reader io.Reader) (*CoapPacket, error) {
	var coapPacket CoapPacket = CoapPacket{0, nil, nil, "", 60, -1, nil}

	bufSingle := make([]byte, 1)
	_, err := io.ReadFull(reader, bufSingle)
	if err != nil {
		return nil, err
	}

	var len byte = (bufSingle[0] & 0xF0) >> 4
	var tklLen byte = bufSingle[0] & 0x0F

	if len == 13 {
		_, err := io.ReadFull(reader, bufSingle)
		if err != nil {
			return nil, err
		}
		len += bufSingle[0]
	}

	var totalCoapSize = len + tklLen + 1
	buf := make([]byte, totalCoapSize)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}

	var index uint8 = 0

	coapPacket.Code = buf[index]
	coapPacket.Token = buf[1:(tklLen + 1)]
	index += tklLen + 1

	//parse options
	var optNum uint8 = 0
	for totalCoapSize > index && buf[index] != 0xFF {
		optDelta := buf[index] >> 4
		optLen := buf[index] & 0x0f
		if optDelta == 13 {
			index++
			optDelta += buf[index]
		}
		if optLen == 13 {
			index++
			optLen += buf[index]
		}

		optNum += optDelta
		optVal := buf[index+1 : index+1+optLen]

		index += 1 + optLen

		switch optNum {
		case 2: //csm
			if coapPacket.Code == CODE_701_CSM {
				csm := Capabilities{readUint32(optVal), false}
				coapPacket.CSM = &csm
			}
		case 4: //csm
			if coapPacket.Code == CODE_701_CSM {
				if coapPacket.CSM != nil {
					coapPacket.CSM.BlockWiseTransfer = true
				} else {
					csm := Capabilities{1152, true}
					coapPacket.CSM = &csm
				}
			}
		case 11: //uri-path
			coapPacket.UriPath += "/" + string(optVal)
		case 12: //content-format
			coapPacket.ContentFormat = int16(optVal[0])
		case 14: //max-age
			coapPacket.MaxAge = readUint32(optVal)
		}
	}

	if totalCoapSize > index && buf[index] == 0xFF {
		coapPacket.Payload = buf[index+1:]
	} else {
		coapPacket.Payload = buf[0:0]
	}

	return &coapPacket, nil
}

func (p CoapPacket) Write(writer io.Writer) error {

	//options
	optBytes := p.writeOptions()

	msgLen := len(optBytes) + len(p.Payload)
	fmt.Printf("len: %d", msgLen)
	if len(p.Payload) > 0 {
		msgLen += 1
	}

	//LEN | TKL
	if msgLen < 13 {
		_, err := writer.Write([]byte{byte(msgLen<<4) + byte(len(p.Token))})
		if err != nil {
			return err
		}
	} else if msgLen < 269 {
		writer.Write([]byte{byte(13<<4) + byte(len(p.Token))})
		writer.Write([]byte{byte(msgLen - 13)})
	} else if msgLen < 65805 {
		writer.Write([]byte{byte(14<<4) + byte(len(p.Token))})
		writer.Write([]byte{byte((msgLen - 269) >> 8)})
		writer.Write([]byte{byte((msgLen - 269) & 0xFF)})
	} else {
		writer.Write([]byte{byte(15<<4) + byte(len(p.Token))})
		writer.Write([]byte{byte((msgLen - 65805) >> 16)})
		writer.Write([]byte{byte((msgLen - 65805) >> 8)})
		writer.Write([]byte{byte((msgLen - 65805) & 0xFF)})
	}

	//Code
	writer.Write([]byte{p.Code})

	//Token
	writer.Write(p.Token)

	//Options
	writer.Write(optBytes)

	//Payload
	if len(p.Payload) > 0 {
		writer.Write([]byte{0xFF})
		writer.Write(p.Payload)
	}

	return nil
}

func (p *CoapPacket) String() string {
	csmString := ""
	if p.CSM != nil {
		csmString = fmt.Sprintf("max-msg-size: %d, block: %t, ", p.CSM.MaxMessageSize, p.CSM.BlockWiseTransfer)
	}

	return fmt.Sprintf("[%s, token:%x, uri-path: %s, cont: %d, max-age: %d, %spayload: %x]", p.StringCode(), p.Token, p.UriPath, p.ContentFormat, p.MaxAge, csmString, p.Payload)
}

func (p *CoapPacket) StringCode() string {
	switch p.Code {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	default:
		return fmt.Sprintf("%d.%02d", p.Code>>5, p.Code&0x1F)

	}

}

func delta(lastOptNum *byte, optionNumber byte) byte {
	delta := optionNumber - *lastOptNum

	*lastOptNum = optionNumber
	return delta
}

func (p *CoapPacket) writeOptions() []byte {
	optWriter := new(bytes.Buffer)
	lastOptNum := byte(0)

	//#2
	if p.CSM != nil {
		p.writeOptionHeaderDynamicSize(optWriter, delta(&lastOptNum, 2), writeUint32(p.CSM.MaxMessageSize))
	}

	//#4
	if p.CSM != nil && p.CSM.BlockWiseTransfer {
		p.writeOptionHeader(optWriter, delta(&lastOptNum, 4), []byte{})
	}

	//#11 uri-path
	if p.UriPath != "" {

		uriPaths := strings.Split(p.UriPath, "/")
		p.writeOptionHeader(optWriter, delta(&lastOptNum, 11), []byte(uriPaths[1]))

		for i := 2; i < len(uriPaths); i++ {
			p.writeOptionHeader(optWriter, 0, []byte(uriPaths[i]))
		}
	}

	//#12 content-format
	if p.ContentFormat >= 0 {
		p.writeOptionHeader(optWriter, delta(&lastOptNum, 12), []byte{byte(p.ContentFormat)})
	}

	//#14 max-age
	if p.MaxAge != 60 {
		p.writeOptionHeaderDynamicSize(optWriter, delta(&lastOptNum, 14), writeUint32(p.MaxAge))
	}

	return optWriter.Bytes()
}

func (p *CoapPacket) writeOptionHeader(optWriter *bytes.Buffer, delta byte, data []byte) {
	size := byte(len(data))
	if delta <= 12 && size <= 12 {
		optWriter.WriteByte(delta<<4 + size)
	} else if delta <= 12 && size > 12 {
		optWriter.WriteByte(delta<<4 + 13)
		optWriter.WriteByte(size - 13)
	} else if delta > 12 && size < 12 {
		optWriter.WriteByte(13<<4 + size)
	} else {
		optWriter.WriteByte(13<<4 + 13)
		optWriter.WriteByte(delta - 13)
		optWriter.WriteByte(size - 13)
	}

	optWriter.Write(data)
}

func (p *CoapPacket) writeOptionHeaderDynamicSize(optWriter *bytes.Buffer, delta byte, data []byte) {
	minData := data

	for minData[0] == 0 {
		minData = minData[1:]
	}

	p.writeOptionHeader(optWriter, delta, minData)
}

func readUint32(data []byte) uint32 {
	size := len(data)

	switch size {
	case 0:
		return 0
	case 1:
		return uint32(data[0])
	case 2:
		return uint32(data[0])<<8 + uint32(data[1])
	case 3:
		return uint32(data[0])<<16 + uint32(data[1])<<8 + uint32(data[2])
	case 4:
		return uint32(data[0])<<24 + uint32(data[1])<<16 + uint32(data[2])<<8 + uint32(data[3])
	default:
		panic("not supported")
	}
}

func writeUint32(data uint32) []byte {
	return []byte{byte(data >> 24), byte(data >> 16), byte(data >> 8), byte(data)}
}
