default: services test

.PHONY: test
test:
	go test ./...

.PHONY: cover
cover:
	PROJECT_DIR=$(shell pwd)  go test -coverprofile=cover.out -coverpkg=./... -v ./...
	@go tool cover -html=cover.out -o cover.html

.PHONY: publish_cover
publish_cover: cover
	go get -d golang.org/x/tools/cmd/cover
	go get github.com/modocache/gover
	go get github.com/mattn/goveralls
	gover
	@goveralls -coverprofile=gover.coverprofile -service=travis-ci -repotoken=$(COVERALLS_TOKEN)

.PHONY: clean
clean:
	@find . -name \.coverprofile -type f -delete
	@rm -f gover.coverprofile

.PHONY: lint
lint: govet golint

.PHONY: govet
govet:
	@echo "Running govet"
	@go vet 

.PHONY: golint
golint: $(GOLINT)
	@echo "Running golint"
	@golint