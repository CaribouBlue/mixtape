FROM golang:1.24 AS build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .
RUN templ generate
RUN go build -v -o ./bin/app ./cmd/server
RUN go build -v -o ./bin/sqlite-setup ./cmd/tools/sqlite/setup

FROM golang:1.24

LABEL org.opencontainers.image.source=https://github.com/CaribouBlue/mixtape
LABEL org.opencontainers.image.description="Mixtape application image"
LABEL org.opencontainers.image.licenses=MIT

ARG PORT=80
ARG APP_DATA_PATH=/var/lib/app

ENV PORT=${PORT}
ENV SERVER_ADDRESS=:${PORT}
ENV APP_DATA_PATH=${APP_DATA_PATH}

COPY --from=build /usr/src/app/bin /usr/local/bin
COPY --from=build /usr/src/app/static ${APP_DATA_PATH}/static

EXPOSE ${PORT}

CMD ["app"]