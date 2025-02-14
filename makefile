.PHONY: start
start:
	go run ./cmd/server

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -file=input.css -xfile=_templ.go templ generate :: npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css :: go run ./cmd/server

.PHONY: sqlite-setup-local
sqlite-setup-local:
	go run ./cmd/tools/sqlite/setup-local

.PHONY: sqlite-teardown-local
sqlite-teardown-local:
	go run ./cmd/tools/sqlite/teardown-local