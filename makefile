.PHONY: start
start:
	go run ./cmd/server

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -file=input.css -xfile=_templ.go templ generate :: npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css :: go run ./cmd/server

.PHONY: sqlite-setup
sqlite-setup:
	go run ./cmd/tools/sqlite/setup

.PHONY: sqlite-teardown
sqlite-teardown:
	go run ./cmd/tools/sqlite/teardown

.PHONY: sqlite-load-test-data
sqlite-load-test-data:
	go run ./cmd/tools/sqlite/load-test-data

.PHONY: sqlite-init-local-dev
sqlite-init-local-dev: 
	make sqlite-teardown 
	make sqlite-setup 
	make sqlite-load-test-data