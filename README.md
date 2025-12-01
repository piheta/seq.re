# seq.re

[![Go Report Card](https://goreportcard.com/badge/github.com/piheta/seq.re)](https://goreportcard.com/report/github.com/piheta/seq.re)
[![Go Test](https://github.com/piheta/seq.re/actions/workflows/test.yml/badge.svg)](https://github.com/piheta/seq.re/actions/workflows/test.yml)
[![Go Lint](https://github.com/piheta/seq.re/actions/workflows/lint.yml/badge.svg)](https://github.com/piheta//actions/workflows/lint.yml)
[![CodeQL](https://github.com/piheta/seq.re/actions/workflows/codeql.yml/badge.svg)](https://github.com/piheta/seq.re/security/code-scanning)

A self-hostable collection of everyday utilities — URL shortening, IP lookup, and secret sharing — without the ads, telemetry, or third-party dependencies.

## Features

- **URL Shortening** - Create short, unique 6-character codes for long URLs with automatic 7-day expiration
- **IP Detection** - Extract client IP addresses with support for proxied requests (X-Forwarded-For, X-Real-IP)
- **Ephemeral Secret Sharing** - Create one time use links for secret sharing
- **Encrypted Badger Database** - Embedded key-value store with automatic TTL-based link expiration
- **CLI Tool** - Command line tool for interacting with the api.

## Server Deployment

### With Database Encryption (AES-256)
```bash
# Generate a random 32-byte (256-bit) encryption key
openssl rand -hex 32

# Use the generated key
docker run -p 8080:8080 \
    -v ./data:/data \
    -e REDIRECT_HOST=https://your-seqre-server.com \
    -e REDIRECT_PORT=:8443 \
    -e BEHIND_PROXY=true \
    -e DB_ENCRYPTION_KEY=your_64_character_hex_key_here \
    piheta/seqre:latest
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIRECT_HOST` | `http://localhost` | Base URL for shortened links |
| `REDIRECT_PORT` | `:8080` | Port suffix for URLs (use `:443` or empty for standard ports) |
| `BEHIND_PROXY` | `false` | Set to `true` when behind Cloudflare/Nginx to trust proxy headers |
| `DB_PATH` | `/data/badger` | Database storage path | Optional: Override the default db path
| `DB_ENCRYPTION_KEY` | - | Optional: 32/48/64 hex chars for AES-128/192/256 encryption |

**Important:** Store the encryption key securely! Without it, your database cannot be decrypted.

## CLI

### Install

```bash
brew tap piheta/seqre
brew install seqre
```

Or download binaries [here](https://github.com/piheta/seq.re/releases)

### Usage

```bash
Usage: seqre <command> [args]
Commands:
  url <URL>              Create a shortened URL
  ip                     Get your IP address
  config set <server>    Set the server URL
  config get             Get the server URL
```

## Roadmap
- Secret sharing
- Fragments for private short urls
