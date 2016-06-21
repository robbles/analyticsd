## analyticsd

A microservice for collecting analytics and periodically rotating them to S3 as JSON logs.

You can log records using any of the following methods:

### 1. Query Params

Request:

```http
GET /?key1=value1&key2=value2 HTTP/1.1
```

Format of S3 log lines:

```json
{"key1": "value1", "key2": "value2"}
```

### 2. POST Body

```http
POST / HTTP/1.1
Content-Type: application/json
Content-Length: 26

{"message": "hello world"}
```

```json
{"message":"hello world"}
```

### 2. Base64-encoded JSON query parameter

```http
GET /track.gif?data=eyJtZXNzYWdlOiAiaGVsbG8gd29ybGQifQ== HTTP/1.1
```

```json
{"message":"hello world"}
```

Note: This endpoint will return a 1x1 transparent GIF as the response. It's
designed to allow embedding of tracking URLs within pixels in webpages and
emails.

### Command-line usage

```
Usage of analyticsd:
  -debug
        Debug mode: log to stderr instead of S3 (default true)
  -aws-region string
        AWS region (default "us-west-1")
  -bucket string
        S3 bucket for storing logs (default "logs")
  -host string
        Host to bind HTTP server on (default "0.0.0.0")
  -key-prefix string
        Prefix for S3 keys
  -logging-dir string
        Directory to store temp log files (default ".")
  -max-log-age duration
        Maximum age logs can reach before rotating to S3 (default 1m0s)
  -max-log-lines int
        Maximum number of lines to log before rotating to S3 (default 100000)
  -num-workers int
        Number of workers uploading logs (default 1)
  -port int
        Port to bind HTTP server on (default 3000)
```
