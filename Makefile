APP_BIN=./bin/api
MAIN_PKG=./cmd/api

.PHONY: build run daemon

build:
	mkdir -p ./bin
	go build -o $(APP_BIN) $(MAIN_PKG)

run: build
	$(APP_BIN)

daemon:
	mkdir -p ./bin
	go run github.com/githubnemo/CompileDaemon@latest \
		--build="go build -o $(APP_BIN) $(MAIN_PKG)" \
		--command="$(APP_BIN)" \
		--exclude-dir=.git \
		--exclude-dir=bin
