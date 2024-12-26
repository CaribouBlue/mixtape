.PHONY: start
start:
	go run ./cmd/server

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run ./cmd/server

.PHONY: setup
setup:
	go run ./cmd/tools/testdata