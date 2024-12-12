.PHONY: start
start:
	go run .

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run main.go

.PHONY: setup
setup:
	go run ./test/testdata/testdata.go