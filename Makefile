# Define the application name
APP_NAME := solo

# Define the Go source directory
SRC_DIR := .

# Define the directory where the built application will be installed
INSTALL_DIR := /usr/local/sbin

# Define the Go build command
BUILD_CMD := go build -o $(APP_NAME) $(SRC_DIR)

# Define the Go install command
INSTALL_CMD := sudo install -m 755 $(APP_NAME) $(INSTALL_DIR)

# Define the clean command
CLEAN_CMD := rm -f $(APP_NAME)

# Default target
.PHONY: all
all: build

# Build target
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	$(BUILD_CMD)
	@echo "Build complete."

# Install target
.PHONY: install
install: build
	sudo systemctl stop solo_client || true
	@echo "Installing $(APP_NAME) to $(INSTALL_DIR)..."
	$(INSTALL_CMD)
	@echo "Install complete."
	sudo systemctl start solo_client || true

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning up..."
	$(CLEAN_CMD)
	@echo "Cleanup complete."
