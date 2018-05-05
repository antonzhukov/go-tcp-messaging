PROJECT = unity_hw
HUB = ${PROJECT}/hub
CLIENT = ${PROJECT}/client

.PHONY: build

build:
	go build -o bin/hub ${HUB}
	go build -o bin/client ${CLIENT}
