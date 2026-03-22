build:
	go build -o bin/backstory ./cmd/backstory
	go build -o bin/backstory-mcp ./cmd/backstory-mcp

build-mcp:
	go build -o bin/backstory-mcp ./cmd/backstory-mcp

test:
	go test ./...

clean:
	rm -rf bin/
