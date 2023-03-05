FROM golang:alpine3.17

USER root

WORKDIR /go/src/app
COPY . .

RUN go get ./...

CMD ["go",  "run", "cmd/main.go"]