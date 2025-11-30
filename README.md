# seq.re

[![Go Report Card](https://goreportcard.com/badge/github.com/piheta/seq.re)](https://goreportcard.com/report/github.com/piheta/seq.re)
[![Go Test](https://github.com/piheta/seq.re/actions/workflows/test.yml/badge.svg)](https://github.com/piheta/seq.re/actions/workflows/test.yml)
[![Go Lint](https://github.com/piheta/seq.re/actions/workflows/lint.yml/badge.svg)](https://github.com/piheta//actions/workflows/lint.yml)
[![CodeQL](https://github.com/piheta/seq.re/actions/workflows/codeql.yml/badge.svg)](https://github.com/piheta/seq.re/security/code-scanning)

A self-hostable Go API service that provides URL shortening and client IP detection. 

## Features

- **URL Shortening** - Create short, unique 6-character codes for long URLs with automatic 7-day expiration
- **IP Detection** - Extract client IP addresses with support for proxied requests (X-Forwarded-For, X-Real-IP)
- **Ephemeral Secret Sharing** - Create one time use links for secret sharing
- **Badger Database** - Embedded key-value store with automatic TTL-based link expiration
- **CLI Tool** - Command line tool for interacting with the api.

## Server Deployment
```bash
docker run -p 8080:8080 \
    -v ./data:/data \
    -e REDIRECT_HOST=https://your-seqre-server.com \
    -e REDIRECT_PORT=:8443 \
    piheta/seqre:latest
```

## CLI Usage
seqre config set <https://your-seqre-server.com:8443> // (Optional) Override default server with your own
seqre config get
seqre ip
seqre url example.com

## Roadmap
- Secret sharing
- Fragments for private short urls
