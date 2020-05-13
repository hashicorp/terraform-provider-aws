.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o godot ./cmd/godot

.PHONY: release
release:
	goreleaser release --rm-dist
