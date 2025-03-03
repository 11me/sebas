# Load environment variables from .env file.
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: run

run:
	@go run .

build:
	@go build -o app .

# Catches any unmatched target and does nothing
%:
	@:

