.PHONY: start
start:
	go run ./cmd/server

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -file=input.css -xfile=_templ.go templ generate :: npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css :: go run ./cmd/server

.PHONY: testdata
testdata:
	go run ./cmd/tools/testdata