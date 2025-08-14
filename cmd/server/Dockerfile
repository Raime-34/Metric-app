FROM golang:1.24.2-alpine AS BUILDER

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/server cmd/server
COPY internal/logger internal/logger
COPY internal/model internal/model
COPY internal/repository internal/repository
COPY internal/server internal/server
COPY internal/filemanager internal/filemanager
COPY internal/utils internal/utils

RUN go build -o server ./cmd/server

CMD ["./server"]