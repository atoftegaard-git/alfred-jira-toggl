PROJECT_NAME := "alfred-jira-toggl"
PKG := "git.netic.dk/users/aht_netic.dk/repos/$(PROJECT_NAME)"
PLIST := "info.plist"
SHELL := /bin/bash

GO111MODULE=on

.EXPORT_ALL_VARIABLES:
.PHONY: all dep lint vet build clean

all: build

dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golangci-lint run --timeout 3m

vet: ## Run go vet
	@go vet ./src

build: dep ## Build the binary file
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o workflow/$(PROJECT_NAME)-amd64 ./src
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o workflow/$(PROJECT_NAME)-arm64 ./src

universal-binary:
	@lipo -create -output workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64
	@rm -f workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

clean: ## Remove previous build
	@rm -f workflow/$(PROJECT_NAME) workflow/$(PROJECT_NAME)-amd64 workflow/$(PROJECT_NAME)-arm64

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

package-alfred:
	@cd ./workflow \
	&& zip --exclude prefs.plist -r ../$(PROJECT_NAME).alfredworkflow ./* \
	&& echo -e "Current version:\n$$(rg -e '[1-9]\.[1-9]\.[1-9]' info.plist)\n" | xargs \
	&& read -p "New version (ex. 1.3.5): " newVersion \
	&& sed -i -e 's,<string>[1-9]\.[1-9]\.[1-9]</string>,<string>'"$${newVersion}"'</string>,g' $(PLIST)
