BIN := dyrwood
SERVICE_LOCATION := /usr/libexec/$(BIN)

all: go-build

install:
	install -o root -g root -m 755 "$(GO_PATH)/bin/$(BIN)" "/usr/local/bin/$(BIN)"

clean: go-clean
	rm -rf "/usr/local/bin/$(BIN)"

include Makedeps/*/*.mk
