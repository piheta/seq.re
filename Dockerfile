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
ENV BEHIND_PROXY=false
ENV DB_PATH=/data/badger
# ENV DB_ENCRYPTION_KEY= (optional: 32/48/64 hex chars for AES-128/192/256). make using `openssl rand -hex 32`

VOLUME ["/data"]

# user nobody
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/seqre-server"]
