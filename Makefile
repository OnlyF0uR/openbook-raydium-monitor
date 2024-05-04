# Define Go commands and flags
GOCMD := go
GOBUILD := $(GOCMD) build
GOMOD := $(GOCMD) mod
GOTEST := $(GOCMD) test

# Build flags
LDFLAGS := -s -w
GOFLAGS := -v

# Define output paths and files
OUTPUT_FOLDER := build
OUTPUT_FILE := $(OUTPUT_FOLDER)/solana-monitor

# Define cleanup command
CLEANCMD := -rm $(OUTPUT_FOLDER) -r

# Adjust output file extension for Windows
ifeq ($(shell go env GOOS),windows)
	OUTPUT_FILE := $(OUTPUT_FILE).exe
	CLEANCMD := -rmdir "$(OUTPUT_FOLDER)" /s /q
endif

# Target: build the project
.PHONY: build
build:
	$(GOBUILD) $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(OUTPUT_FILE) cmd/solana-monitor/main.go

# Target: run tests
.PHONY: test
test:
	$(GOTEST) $(GOFLAGS) ./...

# Target: tidy go.mod and go.sum
.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Target: clean up generated files and folders
.PHONY: clean
clean:
	$(CLEANCMD)
