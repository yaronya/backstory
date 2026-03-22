build:
	go build -o bin/backstory ./cmd/backstory

test:
	go test ./...

clean:
	rm -rf bin/
