FROM golang:1.24.2-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/server cmd/server
COPY internal/logger internal/logger
COPY internal/model internal/model
COPY internal/repository internal/repository
COPY internal/handler internal/handler

RUN go build -o server ./cmd/server

CMD ["./server"]