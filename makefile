.PHONY: build
build:
	templ generate
	npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css
	go build -v -o ./bin/app ./cmd/server

.PHONY: start
start:
	make build
	./bin/app

.PHONY: start-dev
start-dev:
	wgo -file=.go -file=.templ -file=input.css -xfile=_templ.go templ generate :: npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css :: go run ./cmd/server

.PHONY: container-build
container-build:
	docker build -t mixtape .

.PHONY: container-start
container-start:
	make container-build
	docker run --rm --name mixtape -p 8080:80 mixtape

.PHONY: container-teardown
container-teardown:
	docker stop mixtape
	docker rm mixtape

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

.PHONY: docker-stack-deploy
docker-stack-deploy:
	docker stack deploy -c ./compose.yaml mixtape --with-registry-auth 

.PHONY: docker-stack-rm
docker-stack-rm:
	docker stack rm mixtape