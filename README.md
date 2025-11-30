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

## Roadmap
- Secret sharing
- CLI tool
