FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./

COPY web/  ./web

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

EXPOSE 7540
ENV TODO_PASSWORD="12345"

RUN go build -o /app/task_tracker ./main.go

CMD ["./task_tracker"]