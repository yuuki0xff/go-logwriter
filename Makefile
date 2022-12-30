.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test -cover -parallel $(shell nproc) -shuffle on ./...
