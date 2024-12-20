.PHONY: start
start:
	go run ./cmd/main.go

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run ./cmd/main.go

.PHONY: setup
setup:
	go run ./cmd/tools/testdata