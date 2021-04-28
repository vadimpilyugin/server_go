all: server

server: *.go
	go build -o server -v
server_arm: *.go
	GOOS=linux GOARCH=arm64 go build -o $@ -v
server.exe:
	GOOS=windows GOARCH=amd64 go build -o $@ -v
clean:
	rm -f server server.exe server_arm
