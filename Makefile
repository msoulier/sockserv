all: client server chain

client: client.go
	go build client.go

server: server.go
	go build server.go

chain: chain.go
	go build chain.go

clean:
	rm -f client server chain
