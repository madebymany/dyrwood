BIN := dyrwood
SERVICE_LOCATION := /usr/libexec/$(BIN)

all: go-build

install:
	mkdir -p "$(SERVICE_LOCATION)"
	install -o root -g root -m 755 "$(GO_PATH)/bin/$(BIN)" "/usr/local/bin/$(BIN)"
	mkdir -p "/etc/sv/$(BIN)"
	cp -R runit/* "/etc/sv/$(BIN)/"
	ln -s "/etc/sv/$(BIN)" "/etc/service/"

clean: go-clean
	rm -rf "$(SERVICE_LOCATION)" "/usr/local/bin/$(BIN)"

include Makedeps/*/*.mk
