FROM golang:latest

WORKDIR /app/build

COPY . .

RUN go build . -o app

COPY app /app

ENV GIN_MODE=release

ENTRYPOINT [ "/app/app" ]