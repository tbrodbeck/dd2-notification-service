# Define variables (adjust as needed)
PROJECT_NAME := $(shell basename $(PWD))  # Get project name from current directory

# Define targets
.PHONY: all run-local build-container run-container clean help

all: run-local

run-local:
  @echo "Running application locally..."
  # Add your local execution command here (e.g., npm start, python main.py)
  go build main.go && ./main

build-container:
  @echo "Building Docker image..."
  docker build -t $(PROJECT_NAME) .

run-container: build-container
  @echo "Running application in container..."
  docker run -it $(PROJECT_NAME)

clean:
  @echo "Cleaning up..."
  docker stop $(PROJECT_NAME) &>/dev/null
  docker rm $(PROJECT_NAME) &>/dev/null

help:
  @grep -E '^##|^.PHONY:' Makefile | sed -e 's/:.*//' | sort
  @echo "  run-local     : Run the application locally."
  @echo "  build-container: Build the Docker image."
  @echo "  run-container  : Run the application in a container."
  @echo "  clean          : Clean up stopped containers."
  @echo "  help           : Display this help message."