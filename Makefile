BINARY := excaliburl
CMD := ./cmd/excaliburl

LDFLAGS := -s -w

.DEFAULT_GOAL := build

.PHONY: build
build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) $(CMD)

.PHONY: install
install:
	go install -ldflags '$(LDFLAGS)' $(CMD)

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -f $(BINARY)
