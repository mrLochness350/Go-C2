C2_BINARY_NAME=C2
AGENT_BINARY_NAME=Agent

build:
	GOARCH=amd64 GOOS=darwin go build -o bin/${C2_BINARY_NAME}_darwin_amd64 src/c2/c2.go
	GOARCH=amd64 GOOS=linux go build -o bin/${C2_BINARY_NAME}_linux_amd64 src/c2/c2.go
	GOARCH=amd64 GOOS=windows go build -o bin/${C2_BINARY_NAME}_windows_amd64 src/c2/c2.go
	GOARCH=amd64 GOOS=darwin go build -o bin/${AGENT_BINARY_NAME}_darwin_amd64 src/agent/agent.go
	GOARCH=amd64 GOOS=windows go build -o bin/${AGENT_BINARY_NAME}_windows_amd64 src/agent/agent.go
	GOARCH=amd64 GOOS=linux go build -o bin/${AGENT_BINARY_NAME}_linux_amd64 src/agent/agent.go

clean:
	go clean
	rm bin/*


