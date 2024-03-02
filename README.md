# httpreqid

httpreqid is a `net/http` compatible middleware designed for generating a request ID for each request if it doesn't already have one, and propagates it into the request context and response. 

By default, this middleware will look for headers such as `X-Request-ID`, `X-Correlation-ID`, `X-Trace-ID`, `Request-ID`, `Correlation-ID`, and `Trace-ID`.

Additionally, this middleware provides a handler for `log/slog` that automatically adds the request ID to the log record.

## Install

```shell
go get github.com/josestg/httpreqid
```

## Examples

1. [A Simple Example of Using `crypto/rand`](examples/randstr)