# kaedama

kaedama is an M3U8 Proxy service built in Go. It provides a simple API to proxy and cache M3U8 playlist streams.

## Features

- Proxy M3U8 streams with caching support
- CORS enabled for web applications
- Configurable via environment variables
- Health check endpoint

## Quick Start

### Using Go

```bash
# Clone the repository
git clone https://github.com/khatibomar/kaedama.git
cd kaedama

# Run the application
make run
```

The server will start on `http://localhost:4140`.

### Using Docker

```bash
# Build and run with Docker Compose
make docker-up
```

## Usage

### Health Check

```bash
curl http://localhost:4140/health
```

### Proxy a Stream

```bash
curl "http://localhost:4140/proxy?url=https://example.com/stream.m3u8"
```

## Configuration

Configure the application using environment variables:

- `PORT`: Server port (default: 4140)
- `HOST`: Server host (default: 0.0.0.0)
- `ENV`: Environment (default: development)
- `LOG_LEVEL`: Log level (default: debug)
- `CACHE_TTL`: Cache TTL in seconds (default: 300)
- `MAX_CACHE_SIZE`: Max cache size in bytes (default: 104857600)
- `REQUEST_TIMEOUT`: Request timeout in milliseconds (default: 30000)
- `CORS_ORIGINS`: Allowed CORS origins (default: *)

## Development

```bash
# Run tests
make test

# Lint code
make lint
```
