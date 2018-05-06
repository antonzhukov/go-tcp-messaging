PROJECT = unity_hw
HUB = ${PROJECT}/hub
CLIENT = ${PROJECT}/client

.PHONY: get

get:
	dep ensure

.PHONY: build

build:
	go build -o bin/hub ${HUB}
	go build -o bin/client ${CLIENT}
