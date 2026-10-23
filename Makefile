
.PHONY: all
all: build

.PHONY: build
build:
	go build -o bin/warp cmd/warp/main.go

.PHONY: benchmark
benchmark:
	go test -bench=. -count=1 -benchtime=5s -benchmem -run=^#

.PHONY: test
test:
	go test ./...
