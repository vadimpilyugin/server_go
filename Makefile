# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=server
BINARY_UNIX=$(BINARY_NAME)_unix

all: $(BINARY_NAME)
$(BINARY_NAME): *.go
	$(GOBUILD) -o $(BINARY_NAME) -v
server_arm: *.go
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME) -v
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	$(GOGET) github.com/jehiah/go-strftime
	$(GOGET) github.com/go-ini/ini


# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
