all: client server

client: client.go
	go build client.go

server: server.go
	go build server.go

clean:
	rm -f client server
