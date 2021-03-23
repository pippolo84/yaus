all: yaus

yaus:
	go build -o build/yaus cmd/yaus/*.go

.PHONY: lint
lint:
	golint -set_exit_status ./... && go vet ./...

.PHONY: test
test:
	go test -v -race ./...

.PHONY: unit-test
unit-test:
	go test -v -race -short ./...

.PHONY: cover
cover:
	go test -race -covermode=atomic -coverprofile=test/cover.out ./...
	go tool cover -html test/cover.out

.PHONY: bench
bench:
	go test -v -bench=. ./...

.PHONY: doc
doc:
	godoc -http=:8080

.PHONY: clean
clean:
	rm -f build/*
	rm -f test/cover.out
