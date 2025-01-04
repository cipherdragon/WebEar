# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=webear
DIST_DIR=dist

all: test build

build: $(DIST_DIR)
	$(GOBUILD) -o $(DIST_DIR)/$(BINARY_NAME) -v
	cp install.sh $(DIST_DIR)/install.sh
	chmod +x $(DIST_DIR)/install.sh
	cp uninstall.sh $(DIST_DIR)/uninstall.sh

clean:
	$(GOCLEAN)
	rm -rf $(DIST_DIR)

test:
	$(GOTEST) -v ./...

deps:
	$(GOGET) -u ./...

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

.PHONY: all build clean test deps