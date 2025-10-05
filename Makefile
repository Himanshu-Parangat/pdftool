#############################
# Special Variables (Colors)#
#############################

Red=\033[0;31m
Green=\033[0;32m
Yellow=\033[0;33m
Blue=\033[0;34m
Purple=\033[0;35m
Cyan=\033[0;36m
White=\033[0;37m
Gray=\033[0;90m
Reset=\033[0m


######################
# Project Variables  #
######################

APPLICATION=pdftool
BIN_DIR=bin
BINARY=$(BIN_DIR)/$(APPLICATION)

TIMESTAMP=$(shell date +"%Y-%b-%d_%H-%M-%S")
ARCH=$(shell go env GOARCH)
OS=$(shell go env GOOS)

SCRIPTS_DIR := $(PWD)/scripts
REFLEX := $(SCRIPTS_DIR)/reflex
TAILWIND := $(SCRIPTS_DIR)/tailwindcss


#########################
# Miscellaneous Targets #
#########################

.DEFAULT_GOAL := help

.PHONY: % help ensure metadata build run deps clean build-all


#############################
# wildcard for Dead Targets #
#############################

%:
	@printf "\n$(Gray)[Dead]$(Cyan) The target $(Red)'$@'$(Cyan) is not hooked up to anything meaningful.$(Reset)\n"
	@printf "$(Yellow)[Hint]$(Cyan) Try running $(Green)'make help'$(Cyan) for usage information.$(Reset)\n"


###############
# Help Target #
###############

help:
	@printf "\n$(Purple)[Greeting]$(Cyan) This Makefile automates build process for '$(APPLICATION)' $(Reset)\n"
	@printf "\n$(Purple)[Help]$(Cyan) Usage: make <target>$(Reset)\n\n"
	@printf "$(Blue)Available Targets:$(Reset)\n"
	@printf "$(Purple)  help$(Reset)        - Display this help message\n"
	@printf "$(Purple)  ensure$(Reset)      - Ensure the required '$(BIN_DIR)' directory exists\n"
	@printf "$(Purple)  metadata$(Reset)    - Display metadata (timestamp, OS, architecture)\n"
	@printf "$(Purple)  deps$(Reset)        - Resolve and verify Go dependencies\n"
	@printf "$(Purple)  build$(Reset)       - Build the application binary\n"
	@printf "$(Purple)  run$(Reset)         - Run the application\n"
	@printf "$(Purple)  clean$(Reset)       - Remove build artifacts\n\n"
	@printf "$(Red)Example Usage:$(Reset)\n"
	@printf "$(Yellow)  make build-all$(Reset)    - Build for all platforms\n"


######################
# Core Build Targets #
######################

ensure:
	@printf "$(Purple)[Setup]$(Cyan) Ensuring bin directory exists...$(Reset)\n"
	@mkdir -p $(BIN_DIR)
	@printf "$(Green)[Done]$(Reset) Created '$(BIN_DIR)' directory (if missing).\n"

metadata:
	@printf "$(Purple)[Meta]$(Cyan) Timestamp: $(Yellow)$(TIMESTAMP)$(Reset)\n"
	@printf "$(Purple)[Meta]$(Cyan) OS:        $(Yellow)$(OS)$(Reset)\n"
	@printf "$(Purple)[Meta]$(Cyan) Arch:      $(Yellow)$(ARCH)$(Reset)\n"

deps:
	@printf "$(Purple)[GO]$(Blue) Resolving dependencies...$(Reset)\n"
	@go mod tidy     || { printf "$(Red)[Error] go mod tidy failed!$(Reset)\n"; exit 1; }
	@go mod download || { printf "$(Red)[Error] go mod download failed!$(Reset)\n"; exit 1; }
	@go mod verify   || { printf "$(Red)[Error] go mod verify failed!$(Reset)\n"; exit 1; }
	@printf "$(Purple)[GO]$(Green) Dependencies resolved successfully!$(Reset)\n"

build: ensure deps tailwind-build
	@printf "$(Purple)[Build]$(Cyan) Building $(APPLICATION)...$(Reset)\n"
	@go build -o $(BINARY) .
	@printf "$(Green)[Success]$(Reset) Binary built at '$(BINARY)'\n"


install-tools:
	@printf "$(Purple)[Tools]$(Cyan) Installing dev tools...$(Reset)\n"

	@mkdir -p $(SCRIPTS_DIR)

	@printf "$(Blue)[Reflex]$(Reset) Installing reflex...\n"
	@GOBIN=$(SCRIPTS_DIR) go install github.com/cespare/reflex@latest

	@# Detect OS and ARCH for Tailwind
	@OS=$$(go env GOOS); \
	ARCH=$$(go env GOARCH); \
	case "$$OS-$$ARCH" in \
		linux-amd64)   FILE=tailwindcss-linux-x64 ;; \
		linux-arm64)   FILE=tailwindcss-linux-arm64 ;; \
		darwin-amd64)  FILE=tailwindcss-macos-x64 ;; \
		darwin-arm64)  FILE=tailwindcss-macos-arm64 ;; \
		windows-amd64) FILE=tailwindcss-windows-x64.exe ;; \
		windows-arm64) FILE=tailwindcss-windows-arm64.exe ;; \
		*386) \
			printf  "$(Red)[Error] Tailwind CLI binary is only available for 64-bit systems.$(Reset)\n"; \
			printf "$(Blue) [Run] npm install -D tailwindcss$(Reset)\n"; \
			exit 0 ;; \
		*) printf  "$(Red)[Error] Unsupported platform ($$OS-$$ARCH).$(Reset)\n" && exit 1 ;; \
	esac; \
	URL="https://github.com/tailwindlabs/tailwindcss/releases/latest/download/$$FILE"; \
	printf "$(Blue)[Tailwind]$(Reset) Downloading $$FILE...\n"; \
	curl -sL -o $(TAILWIND) "$$URL"; \
	chmod +x $(TAILWIND)

	@printf "$(Green)[Success]$(Reset) Tools installed in $(SCRIPTS_DIR)\n"


tailwind-dev:
	@printf "$(Purple)[Tailwind]$(Cyan) Running in dev mode...$(Reset)\n"
	@./scripts/tailwindcss -i build/tailwindcss/input.css -o static/styles/main.css --watch


tailwind-build:
	@printf "$(Purple)[Tailwind]$(Cyan) Building production CSS...$(Reset)\n"
	@./scripts/tailwindcss -i build/tailwindcss/input.css -o static/styles/main.css --minify
	@printf "$(Green)[Success]$(Reset) Tailwind CSS built at 'static/styles.css'\n"


run: build
	@printf "$(Purple)[Run]$(Cyan) Executing $(APPLICATION)...$(Reset)\n\n"
	@./$(BINARY)

dev: 
	@printf "$(Purple)[Dev]$(Cyan) Starting live-reload server with reflex...$(Reset)\n"
	@$(REFLEX) -r '(\.go$$|\.html$$|\.js$$|\.css$$)' -s -- sh -c "go run ."


clean:
	@printf "$(Purple)[Clean]$(Cyan) Removing build artifacts...$(Reset)\n"
	@rm -rf $(BIN_DIR)
	@printf "$(Green)[Done]$(Reset) Build artifacts removed.\n"

build-all: ensure tailwind-build
	@printf "$(Purple)[Build-All]$(Cyan) Building for all platforms...$(Reset)\n"
	@GOOS=linux   GOARCH=amd64 go build -o $(BIN_DIR)/$(APPLICATION)-linux-amd64 .
	@GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APPLICATION)-windows-amd64.exe .
	@GOOS=darwin  GOARCH=amd64 go build -o $(BIN_DIR)/$(APPLICATION)-darwin-amd64 .
	@printf "$(Green)[Success]$(Reset) All binaries are in '$(BIN_DIR)'\n"

