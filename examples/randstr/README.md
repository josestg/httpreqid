# A Simple Example of Using `crypto/rand`

### Run the demo server
```shell
go run examples/randstr/main.go
```

server log:

```shell
{"time":"2024-03-02T16:10:28.134015+07:00","level":"INFO","msg":"server is listening","addr":":8080"}
```


### Send a request with no request ID provided in the header.

```shell
curl http://localhost:8080/ping -i 
```

server log:
```shell
{"time":"2024-03-02T16:11:53.876167+07:00","level":"INFO","msg":"ping requested","request_id":"8d08c3a430ebefc36776594bef3d37ec"}
```

response:

```shell
HTTP/1.1 200 OK
X-Request-Id: 8d08c3a430ebefc36776594bef3d37ec
Date: Sat, 02 Mar 2024 09:11:53 GMT
Content-Length: 51
Content-Type: text/plain; charset=utf-8

PONG! request id "8d08c3a430ebefc36776594bef3d37ec"
```

### Send a request with the `X-Request-ID` header set.

```shell
curl -H 'X-Request-ID: foo' http://localhost:8080/ping -i 
```
server log:

```shell
{"time":"2024-03-02T16:15:15.198534+07:00","level":"INFO","msg":"ping requested","request_id":"foo"}
```

response:

```shell
HTTP/1.1 200 OK
X-Request-Id: foo
Date: Sat, 02 Mar 2024 09:15:15 GMT
Content-Length: 22
Content-Type: text/plain; charset=utf-8

PONG! request id "foo"
```

### Send a request with the `X-Correlation-ID` header set.

```shell
curl -H 'X-Correlation-ID: bar' http://localhost:8080/ping -i
```

server log:

```shell
{"time":"2024-03-02T16:17:44.951151+07:00","level":"INFO","msg":"ping requested","request_id":"bar"}
```

response:

```shell
HTTP/1.1 200 OK
X-Correlation-Id: bar
Date: Sat, 02 Mar 2024 09:17:44 GMT
Content-Length: 22
Content-Type: text/plain; charset=utf-8

PONG! request id "bar"
```