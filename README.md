# Go CoAP over TCP Library

[![Build Status](https://api.travis-ci.org/szymex/go-coap-tcp.svg)](http://travis-ci.org/szymex/go-coap-tcp)
[![codecov](https://codecov.io/gh/szymex/go-coap-tcp/branch/master/graph/badge.svg)](https://codecov.io/gh/szymex/go-coap-tcp)

Implements [RFC-8323](https://tools.ietf.org/html/rfc8323) - CoAP over TCP:
  - server and client
  - simple request/response
  - *[TODO] observations*
  - *[TODO] TLS integration*
  - *[TODO] WebSocket support*


## Example server

Example server listens on default port (5683). It exposes resources:
    
    /time
    /my-ip
    /rfc8323
    /tmp
    /slow
    
### Build and run

    make build
    ./bin/example-server
    
## Simple command client client

### Build and run

    make build
    ./bin/coap-cli

### Usage

```
Usage: coap-cli [options...] <GET|PUT|POST|DELETE|PING> <url> [payload]
Options:
  -cf int
        content format:
          0 - text/plain
          41 - application/xml
          42 - application/octet-stream
          50 - application/json
         (default -1)
  -max-age int
        max age in seconds (default 60)
```


### Examples (with running example server):

```bash
./bin/coap-cli GET coap://localhost:5683/time

./bin/coap-cli POST localhost/tmp "test"

./bin/coap-cli GET localhost/tmp

./bin/coap-cli -max-age=3600 -cf=50 PUT localhost/tmp "{'test': 1234}"
```
    
## License

Apache License, Version 2.0 
