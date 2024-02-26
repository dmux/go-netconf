GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
BINARY_NAME = go-netconf
VERSION = 1.0.0

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

release:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_$(VERSION)_linux_amd64

.PHONY: all build clean run release