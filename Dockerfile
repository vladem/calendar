FROM golang:alpine3.17

RUN adduser --disabled-password --gecos '' api
USER api

WORKDIR /go/src/app
COPY . .

RUN go get ./cmd
RUN go get ./service

CMD ["go",  "run", "cmd/main.go"]