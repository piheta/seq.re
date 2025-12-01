# seq.re

[![Go Report Card](https://goreportcard.com/badge/github.com/piheta/seq.re)](https://goreportcard.com/report/github.com/piheta/seq.re)
[![Go Test](https://github.com/piheta/seq.re/actions/workflows/test.yml/badge.svg)](https://github.com/piheta/seq.re/actions/workflows/test.yml)
[![Go Lint](https://github.com/piheta/seq.re/actions/workflows/lint.yml/badge.svg)](https://github.com/piheta//actions/workflows/lint.yml)
[![CodeQL](https://github.com/piheta/seq.re/actions/workflows/codeql.yml/badge.svg)](https://github.com/piheta/seq.re/security/code-scanning)

A self-hostable collection of everyday utilities I got tired of using through ad-riddled, telemetry-heavy, closed-source services like TinyURL, whatismyipaddress.com, and others.

## Features

- **URL Shortening** - Create short, unique 6-character codes for long URLs with automatic 7-day expiration
- **IP Detection** - Extract client IP addresses with support for proxied requests (X-Forwarded-For, X-Real-IP)
- **Ephemeral Secret Sharing** - Create one time use links for secret sharing
- **Badger Database** - Embedded key-value store with automatic TTL-based link expiration
- **CLI Tool** - Command line tool for interacting with the api.

## Server Deployment

### Direct Deployment
```bash
docker run -p 8080:8080 \
    -v ./data:/data \
    -e REDIRECT_HOST=https://your-seqre-server.com \
    -e REDIRECT_PORT=:8443 \
    piheta/seqre:latest
```

### Behind Reverse Proxy (Cloudflare, Nginx, etc.)
```bash
docker run -p 8080:8080 \
    -v ./data:/data \
    -e REDIRECT_HOST=https://your-seqre-server.com \
    -e REDIRECT_PORT=:8443 \
    -e BEHIND_PROXY=true \
    piheta/seqre:latest
```

**Note:** Set `BEHIND_PROXY=true` when deploying behind Cloudflare, Nginx, or any reverse proxy to correctly detect client IP addresses.

## CLI Usage

Configure the server URL (optional, defaults to `https://seq.re`):
```bash
seqre config set https://your-seqre-server.com:8443
```

Get the configured server URL:
```bash
seqre config get
```

Get your IP Public address:
```bash
seqre ip
```

Create a shortened URL:
```bash
seqre url example.com
```

## Roadmap
- Secret sharing
- Fragments for private short urls
