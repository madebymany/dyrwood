GO_PATH:=/gopath
GO_REMOTE:=$(shell git remote -v | awk '/^origin[ \t]/{print $$2; exit}')
GO_SOURCE:=$(shell pwd)

go-setup-vars:
	$(eval GO_DIR = $(subst git@github.com:, github.com/, $(GO_REMOTE)))
	$(eval GO_DIR = $(subst .git, , $(GO_DIR)))
	$(eval GO_DIR = $(strip $(GO_DIR)))
	$(eval GO_BUILD_DIR = $(GO_PATH)/src/$(GO_DIR))

$(GO_PATH):
	mkdir -p "${GO_PATH}"
	cd "${GO_PATH}" && mkdir -p "pkg bin src"
	mkdir -p "$(GO_BUILD_DIR)" 
	cp -R $(GO_SOURCE)/* "$(GO_BUILD_DIR)"

.PHONY: go-build
go-build: go-setup-vars $(GO_PATH)
	export GOPATH=$(GO_PATH) && cd $(GO_BUILD_DIR) && go get -d -t -v ./...
	export GOPATH=$(GO_PATH) && cd $(GO_BUILD_DIR) && go test ./...
	export GOPATH=$(GO_PATH) &&cd $(GO_BUILD_DIR) && go install ./...

go-clean:
	rm -rf $(GO_PATH)

