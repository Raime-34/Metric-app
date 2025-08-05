FROM golang:1.24.2-alpine AS BUILDER

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/agent cmd/agent
COPY internal/agent internal/agent
COPY internal/logger internal/logger
COPY internal/model internal/model
COPY internal/repository internal/repository
COPY internal/zip internal/zip

RUN go build -o agent ./cmd/agent

CMD ["./agent", "-a", "server:8080"]