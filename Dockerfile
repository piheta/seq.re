# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o seqre-server ./cmd/server

FROM scratch

COPY --from=builder /app/seqre-server /seqre-server

ENV REDIRECT_HOST=http://localhost
ENV REDIRECT_PORT=:8080
ENV DB_PATH=/data/badger

VOLUME ["/data"]

# user nobody
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/seqre-server"]
