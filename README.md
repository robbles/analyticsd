## analyticsd

A small daemon for collecting analytics and periodically rotating them to S3 as JSON logs.

You can log records using any of the following methods:

### 1. Query Params

Request:

```http
GET /track/?key1=value1&key2=value2 HTTP/1.1
```

Format of S3 log lines:

```json
{"key1": "value1", "key2": "value2"}
```

Note: You can also use /track.gif as the path to receive a 1x1 transparent GIF as the response.

### 2. JSON Body

```http
POST /track/ HTTP/1.1
Content-Type: application/json
Content-Length: 26

{"message": "hello world"}
```

```json
{"message":"hello world"}
```

### 2. Base64-encoded JSON query parameter

```http
GET /track/base64/?data=eyJtZXNzYWdlOiAiaGVsbG8gd29ybGQifQ== HTTP/1.1
```

```json
{"message":"hello world"}
```

### Command-line usage

```
Usage of analyticsd:
  -aws-region="us-west-1": AWS region
  -bucket="logs": S3 bucket for storing logs
  -debug=false: Debug mode: log to stderr instead of S3
  -host="0.0.0.0": Host to bind HTTP server on
  -key-prefix="": Prefix for S3 keys
  -logging-dir=".": Directory to store temp log files
  -num-workers=1: Number of workers uploading logs
  -port=3000: Port to bind HTTP server on
```
