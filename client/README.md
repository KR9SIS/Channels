# client

## WebRPC Steps

### STUN Setup

When a peer wants to set up a connection they reach out to the STUN server
to acquire their own public IP address.

### Signalling Server

Before clients can communicate with each other they need to connect
to a shared signalling server.

1. Connect to the signalling server with own Public IP address, port and SDP info
2. Signalling server broadcasts this information to all other clients it knows about.

### P2P through ICE

After clients learn about each others existence and public ip addresses

1. Use ICE to connect to each other.
2. Iterate ICE connection candidates and start streaming data
   as soon as a possible candidate is found.
3. Continue exchanging all possible candidates even after data has started to stream
   in hopes of finding the best one and keeping all other possible candidates in
   mind if the network situation changes.
