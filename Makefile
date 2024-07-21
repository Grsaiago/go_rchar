# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./cmd/rchat/
BINARY_NAME := rchat
PWD := $(shell pwd)

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## build: build the application
.PHONY: build
build:
	@go build -o=/tmp/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## run: run the  application
.PHONY: run
run: build
	@/tmp/bin/${BINARY_NAME}

## dev: up on development environment/refresh app
.PHONY: dev
dev:
	@go build -o=${PWD}/dev/services/app/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## up: up the whole application on a compose
.PHONY: up
up: 
	@echo "TODO, mas tem compose up pra subir db e redis"

## down: down the whole application on a compose
.PHONY: down
down:
	@echo "TODO, mas tem compose down pra descer db e redis"

