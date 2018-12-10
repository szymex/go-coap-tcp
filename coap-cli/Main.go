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
	"errors"
	"flag"
	"fmt"
	"github.com/szymex/go-coap-tcp/coap"
	"net/url"
	"os"
	"strings"
)

func main() {
	//go run ./coap-cli GET coap://localhost:5683/time
	//go run ./coap-cli GET localhost/time

	var contentFormat = flag.Int("cf", -1, "content format:\n  0 - text/plain\n  41 - application/xml\n  42 - application/octet-stream\n  50 - application/json\n")
	var maxAge = flag.Int("max-age", 60, "max age in seconds")
	flag.Parse()

	if flag.NArg() < 2 {
		printUsage()
		return
	}

	payload := []byte(strings.Join(flag.Args()[2:], " "))
	method := parseMethod(flag.Arg(0))
	uri := parseUri(flag.Arg(1))

	req := coap.NewCoapPacket(method, []byte{}, payload)
	req.UriPath = uri.Path
	req.ContentFormat = int16(*contentFormat)
	req.MaxAge = uint32(*maxAge)

	client, err := coap.Connect(uri.Host)
	if err != nil {
		exit(err)
	}

	resp, err := client.InvokeCoap(req)
	if err != nil {
		exit(err)
	}

	if len(resp.Payload) > 0 {
		fmt.Println("")
		fmt.Println(string(resp.Payload))
	}
}

func printUsage() {
	fmt.Println("Usage: coap-cli [options...] <GET|PUT|POST|DELETE|PING> <url> [payload]")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("Example:")
	fmt.Println("  coap-cli GET coap://localhost:5683/time")
	fmt.Println("  coap-cli PUT coap://localhost:5683/tmp Lorem ipsum")
}

func parseUri(uri string) *url.URL {
	u, err := url.Parse(uri)
	if err != nil {
		exit(err)
	}
	if u.Port() == "" {
		u.Host = u.Host + ":5683"
	}
	return u
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func parseMethod(strMethod string) uint8 {
	switch strMethod {
	case "GET":
		return coap.GET
	case "POST":
		return coap.POST
	case "PUT":
		return coap.PUT
	case "DELETE", "DEL":
		return coap.DELETE
	case "PING":
		return coap.CODE_702_PING
	default:
		exit(errors.New("Malformed method name: " + strMethod))
		return 0
	}
}
