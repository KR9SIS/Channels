# client

## WebRPC Steps

### STUN Setup

When a peer wants to set up a connection they reach out to the STUN server to acquire their own public IP address.

### Signalling Server

Before clients can communicate with each other they need to connect to a shared signalling server. They connect to the signalling server with their own Public IP address, port and SDP info. The signalling server then broadcasts this information to all other clients it knows about.

### P2P through ICE

After clients learn about each others existence and public ip addresses they use ICE to connect to each other. They go through each ICE connection candidate and then start streaming data between them as soon as a possible candidate is found. However they continue exchanging all possible candidates even after data has started to stream in hopes of finding the best one and keeping all other possible candidates in mind if the network situation changes.
