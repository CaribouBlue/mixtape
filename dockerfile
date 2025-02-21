FROM golang:1.24

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN go build -v -o /usr/local/bin/app ./cmd/server

ENV DB_PATH /data/db.sqlite3
ENV SERVER_ADDRESS :80
EXPOSE 80

RUN go run ./cmd/tools/sqlite/setup

CMD ["app"]