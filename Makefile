.PHONY: build build-agent clean all

build:
	go build -o bin/nexus.exe ./cmd/nexus

build-agent:
	cd agent && go build -o ../bin/nexus-agent.exe ./cmd/agent

clean:
	rm -rf bin/

all: build build-agent