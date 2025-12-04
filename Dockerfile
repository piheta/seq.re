# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o seqre-server ./cmd/server

FROM scratch

COPY --from=builder /app/seqre-server /seqre-server
COPY --from=builder /app/web /web

ENV REDIRECT_HOST=http://localhost
ENV REDIRECT_PORT=:8080
ENV BEHIND_PROXY=false
ENV DATA_PATH=/data/seqre
# ENV DB_ENCRYPTION_KEY= (optional: 32/48/64 hex chars for AES-128/192/256). make using `openssl rand -hex 32`
# ENV CONTACT_EMAIL= (optional: contact email for the frontend footer)

VOLUME ["/data"]

# user nobody
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/seqre-server"]
