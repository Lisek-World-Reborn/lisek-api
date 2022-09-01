FROM golang:latest as build

WORKDIR /app/build

COPY . .

RUN go mod download

RUN go build -o lisek-api .

FROM debian:10-slim as production

WORKDIR /app

COPY --from=build /app/build/lisek-api .

ENV GIN_MODE=release

ENTRYPOINT [ "/app/lisek-api" ]