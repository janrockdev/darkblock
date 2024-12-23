build:
	go build -o ./bin/darkblock

run: build
	./bin/darkblock

run3: build
	./bin/darkblock -port=:3000

run4: build
	./bin/darkblock -port=:4000

run5: build
	./bin/darkblock -port=:5000

run6: build
	./bin/darkblock -port=:6000

tx:
	go run ./client/client.go

test:
	go test ./...

testv:
	go test -v ./...

pb:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

.PHONY: proto