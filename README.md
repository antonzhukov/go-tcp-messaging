Description
===========
This repo contains hub and client in separate packages.
The concept of messaging is based on protobuf format.
The client is supplemented with simple cli api.
The technical requirements are [here](task/README.md)

Installation
============
You can build both client and hub with:

    make get
    make build

Usage
=====
Start hub

    ./bin/hub

Start clients

    ./bin/client

Type `help` to get list of available commands

Assumptions
===========
1. User authentication. To authenticate users,
I've chosen naive approach with simple integer counter,
every new user gets an incremented integer as ID
2. Messaging fashion. Client is sending requests and receiving responses
and relays from server. Whenever client is sending request it blocks on
responseChan and waits for asynchronous handler to write response
in the responseChan.
3. I've chosen protobuf as the most convenient protocol for network communication.
Based on benchmarks from https://github.com/alecthomas/go_serialization_benchmarks
github.com/gogo/protobuf seems like the most advanced in terms of
memory allocations and speed.
4. There's request timeout implemented in the code to
terminate the unsuccessful requests.

Trade-Offs
==========
1. In the implementation there's no real multiplexing.
Blocking on responseChan is somewhat fragile, since we are receiving
messages asynchronously and don't know really if the received response
is actually for our certain request. Though synchronous request send
saves us from this issue. Multiplexing could be solved with
streams, e.g. with: https://github.com/hashicorp/yamux
2. Both client logging and CLI are using stdout for simplicity. 
