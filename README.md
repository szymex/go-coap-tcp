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

## License

Apache License, Version 2.0 
