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

const rfc8323 = `
Internet Engineering Task Force (IETF)                        C. Bormann
Request for Comments: 8323                       Universitaet Bremen TZI
Updates: 7641, 7959                                             S. Lemay
Category: Standards Track                             Zebra Technologies
ISSN: 2070-1721                                            H. Tschofenig
                                                                ARM Ltd.
                                                               K. Hartke
                                                 Universitaet Bremen TZI
                                                           B. Silverajan
                                        Tampere University of Technology
                                                          B. Raymor, Ed.
                                                           February 2018


 CoAP (Constrained Application Protocol) over TCP, TLS, and WebSockets

Abstract

   The Constrained Application Protocol (CoAP), although inspired by
   HTTP, was designed to use UDP instead of TCP.  The message layer of
   CoAP over UDP includes support for reliable delivery, simple
   congestion control, and flow control.

   Some environments benefit from the availability of CoAP carried over
   reliable transports such as TCP or Transport Layer Security (TLS).
   This document outlines the changes required to use CoAP over TCP,
   TLS, and WebSockets transports.  It also formally updates RFC 7641
   for use with these transports and RFC 7959 to enable the use of
   larger messages over a reliable transport.

Status of This Memo

   This is an Internet Standards Track document.

   This document is a product of the Internet Engineering Task Force
   (IETF).  It represents the consensus of the IETF community.  It has
   received public review and has been approved for publication by the
   Internet Engineering Steering Group (IESG).  Further information on
   Internet Standards is available in Section 2 of RFC 7841.

   Information about the current status of this document, any errata,
   and how to provide feedback on it may be obtained at
   https://www.rfc-editor.org/info/rfc8323.








Bormann, et al.              Standards Track                    [Page 1]

RFC 8323         TCP/TLS/WebSockets Transports for CoAP    February 2018


Copyright Notice

   Copyright (c) 2018 IETF Trust and the persons identified as the
   document authors.  All rights reserved.

   This document is subject to BCP 78 and the IETF Trust's Legal
   Provisions Relating to IETF Documents
   (https://trustee.ietf.org/license-info) in effect on the date of
   publication of this document.  Please review these documents
   carefully, as they describe your rights and restrictions with respect
   to this document.  Code Components extracted from this document must
   include Simplified BSD License text as described in Section 4.e of
   the Trust Legal Provisions and are provided without warranty as
   described in the Simplified BSD License.

Table of Contents

   1. Introduction ....................................................3
   2. Conventions and Terminology .....................................6
   3. CoAP over TCP ...................................................7
      3.1. Messaging Model ............................................7
      3.2. Message Format .............................................9
      3.3. Message Transmission ......................................11
      3.4. Connection Health .........................................12
   4. CoAP over WebSockets ...........................................13
      4.1. Opening Handshake .........................................15
      4.2. Message Format ............................................15
      4.3. Message Transmission ......................................16
      4.4. Connection Health .........................................17
   5. Signaling ......................................................17
      5.1. Signaling Codes ...........................................17
      5.2. Signaling Option Numbers ..................................18
      5.3. Capabilities and Settings Messages (CSMs) .................18
      5.4. Ping and Pong Messages ....................................20
      5.5. Release Messages ..........................................21
      5.6. Abort Messages ............................................23
      5.7. Signaling Examples ........................................24
   6. Block-Wise Transfer and Reliable Transports ....................25
      6.1. Example: GET with BERT Blocks .............................27
      6.2. Example: PUT with BERT Blocks .............................27
   7. Observing Resources over Reliable Transports ...................28
      7.1. Notifications and Reordering ..............................28
      7.2. Transmission and Acknowledgments ..........................28
      7.3. Freshness .................................................28
      7.4. Cancellation ..............................................29






Bormann, et al.              Standards Track                    [Page 2]

RFC 8323         TCP/TLS/WebSockets Transports for CoAP    February 2018


   8. CoAP over Reliable Transport URIs ..............................29
      8.1. coap+tcp URI Scheme .......................................30
      8.2. coaps+tcp URI Scheme ......................................31
      8.3. coap+ws URI Scheme ........................................32
      8.4. coaps+ws URI Scheme .......................................33
      8.5. Uri-Host and Uri-Port Options .............................33
      8.6. Decomposing URIs into Options .............................34
      8.7. Composing URIs from Options ...............................35
   9. Securing CoAP ..................................................35
      9.1. TLS Binding for CoAP over TCP .............................36
      9.2. TLS Usage for CoAP over WebSockets ........................37
   10. Security Considerations .......................................37
      10.1. Signaling Messages .......................................37
   11. IANA Considerations ...........................................38
      11.1. Signaling Codes ..........................................38
      11.2. CoAP Signaling Option Numbers Registry ...................38
      11.3. Service Name and Port Number Registration ................40
      11.4. Secure Service Name and Port Number Registration .........40
      11.5. URI Scheme Registration ..................................41
      11.6. Well-Known URI Suffix Registration .......................43
      11.7. ALPN Protocol Identifier .................................44
      11.8. WebSocket Subprotocol Registration .......................44
      11.9. CoAP Option Numbers Registry .............................44
   12. References ....................................................45
      12.1. Normative References .....................................45
      12.2. Informative References ...................................47
   Appendix A. Examples of CoAP over WebSockets ......................49
   Acknowledgments ...................................................52
   Contributors ......................................................52
   Authors' Addresses ................................................53

1.  Introduction

   The Constrained Application Protocol (CoAP) [RFC7252] was designed
   for Internet of Things (IoT) deployments, assuming that UDP [RFC768]
   can be used unimpeded as can the Datagram Transport Layer Security
   (DTLS) protocol [RFC6347] over UDP.  The use of CoAP over UDP is
   focused on simplicity, has a low code footprint, and has a small
   over-the-wire message size.

   The primary reason for introducing CoAP over TCP [RFC793] and TLS
   [RFC5246] is that some networks do not forward UDP packets.  Complete
   blocking of UDP happens in between about 2% and 4% of terrestrial
   access networks, according to [EK2016].  UDP impairment is especially
   concentrated in enterprise networks and networks in geographic
   regions with otherwise challenged connectivity.  Some networks also





Bormann, et al.              Standards Track                    [Page 3]

RFC 8323         TCP/TLS/WebSockets Transports for CoAP    February 2018


   rate-limit UDP traffic, as reported in [BK2015], and deployment
   investigations related to the standardization of Quick UDP Internet
   Connections (QUIC) revealed numbers around 0.3% [SW2016].

   The introduction of CoAP over TCP also leads to some additional
   effects that may be desirable in a specific deployment:

   o  Where NATs are present along the communication path, CoAP over TCP
      leads to different NAT traversal behavior than CoAP over UDP.
      NATs often calculate expiration timers based on the
      transport-layer protocol being used by application protocols.
      Many NATs maintain TCP-based NAT bindings for longer periods based
      on the assumption that a transport-layer protocol, such as TCP,
      offers additional information about the session lifecycle.  UDP,
      on the other hand, does not provide such information to a NAT and
      timeouts tend to be much shorter [HomeGateway].  According to
      [HomeGateway], the mean for TCP and UDP NAT binding timeouts is
      386 minutes (TCP) and 160 seconds (UDP).  Shorter timeout values
      require keepalive messages to be sent more frequently.  Hence, the
      use of CoAP over TCP requires less-frequent transmission of
      keepalive messages.

   o  TCP utilizes mechanisms for congestion control and flow control
      that are more sophisticated than the default mechanisms provided
      by CoAP over UDP; these TCP mechanisms are useful for the transfer
      of larger payloads.  (However, work is ongoing to add advanced
      congestion control to CoAP over UDP as well; see [CoCoA].)

   Note that the use of CoAP over UDP (and CoAP over DTLS over UDP) is
   still the recommended transport for use in constrained node networks,
   particularly when used in concert with block-wise transfer.  CoAP
   over TCP is applicable for those cases where the networking
   infrastructure leaves no other choice.  The use of CoAP over TCP
   leads to a larger code size, more round trips, increased RAM
   requirements, and larger packet sizes.  Developers implementing CoAP
   over TCP are encouraged to consult [TCP-in-IoT] for guidance on
   low-footprint TCP implementations for IoT devices.

   Standards based on CoAP, such as Lightweight Machine to Machine
   [LWM2M], currently use CoAP over UDP as a transport; adding support
   for CoAP over TCP enables them to address the issues above for
   specific deployments and to protect investments in existing CoAP
   implementations and deployments.

   Although HTTP/2 could also potentially address the need for
   enterprise firewall traversal, there would be additional costs and
   delays introduced by such a transition from CoAP to HTTP/2.
   Currently, there are also fewer HTTP/2 implementations available for



Bormann, et al.              Standards Track                    [Page 4]

RFC 8323         TCP/TLS/WebSockets Transports for CoAP    February 2018


   constrained devices in comparison to CoAP.  Since CoAP also supports
   group communication using IP-layer multicast and unreliable
   communication, IoT devices would have to support HTTP/2 in addition
   to CoAP.

   Furthermore, CoAP may be integrated into a web environment where the
   front end uses CoAP over UDP from IoT devices to a cloud
   infrastructure and then CoAP over TCP between the back-end services.
   A TCP-to-UDP gateway can be used at the cloud boundary to communicate
   with the UDP-based IoT device.

   Finally, CoAP applications running inside a web browser may be
   without access to connectivity other than HTTP.  In this case, the
   WebSocket Protocol [RFC6455] may be used to transport CoAP requests
   and responses, as opposed to cross-proxying them via HTTP to an
   HTTP-to-CoAP cross-proxy.  This preserves the functionality of CoAP
   without translation -- in particular, the Observe Option [RFC7641].

   To address the above-mentioned deployment requirements, this document
   defines how to transport CoAP over TCP, CoAP over TLS, and CoAP over
   WebSockets.  For these cases, the reliability offered by the
   transport protocol subsumes the reliability functions of the message
   layer used for CoAP over UDP.  (Note that for both a reliable
   transport and the message layer for CoAP over UDP, the reliability
   offered is per transport hop: where proxies -- see Sections 5.7 and
   10 of [RFC7252] -- are involved, that layer's reliability function
   does not extend end to end.)  Figure 1 illustrates the layering:

     +--------------------------------+
     |          Application           |
     +--------------------------------+
     +--------------------------------+
     |  Requests/Responses/Signaling  |  CoAP (RFC 7252) / This Document
     |--------------------------------|
     |        Message Framing         |  This Document
     +--------------------------------+
     |      Reliable Transport        |
     +--------------------------------+

            Figure 1: Layering of CoAP over Reliable Transports

   This document specifies how to access resources using CoAP requests
   and responses over the TCP, TLS, and WebSocket protocols.  This
   allows connectivity-limited applications to obtain end-to-end CoAP
   connectivity either (1) by communicating CoAP directly with a CoAP
   server accessible over a TCP, TLS, or WebSocket connection or (2) via
   a CoAP intermediary that proxies CoAP requests and responses between
   different transports, such as between WebSockets and UDP.



Bormann, et al.              Standards Track                    [Page 5]
`
