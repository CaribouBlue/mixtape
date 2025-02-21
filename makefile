.PHONY: build
build:
	go build -v -o ./bin/app ./cmd/server

.PHONY: start
start:
	make build
	./bin/app

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -file=input.css -xfile=_templ.go templ generate :: npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css :: go run ./cmd/server

.PHONY: build-container
build-container:
	docker build -t top-spot .

.PHONY: start-container
start-container:
	make build-container
	docker run --rm --name top-spot -p 8080:80 top-spot

.PHONY: teardown-container
teardown-container:
	docker stop top-spot
	docker rm top-spot

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