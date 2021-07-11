

# The Gateway for Apron Project

Apron gateway handles the authentication and authorization of user who want to access revisited services,
and forward the request to service's API if check passed.
The gateway currently can support *http*, *https*, *ws*, *wss* schema.
Besides forwarding the request, the gateway also provides ability to aggregate user request and generate usage report.
With apron node, this report will be published to chain and can be checked by everyone for audition.

After starting, the gateway will serve on two port,
one is for admin API and the other is for request forward.
The default server address for admin API is *127.0.0.1:8082*
while the default request forward address is *0.0.0.0:8080*.

This document how to setup the service and forward a request

> For security factor, the admin API can only be accessed via local node by default, 
> but it can be configured via environment variable. The detail will be shown in setup section below.

## Build

### Docker

```shell
$ docker build -t apron/gateway .
```

### Standalone

> Note: `protoc` and `protoc-gen-go` are required if want to regenerate protobuf model.

```shell
$ make gen  # Generate protobuf model
$ make build # Build gateway binary
```

## Environment setup

### Docker compose

The `docker-compose.yml` file shipped with project build a cluster with a redis server and a httpbin server.
The redis server is used for gateway service for storing service and user data,
while the httpbin service is used for demo request forward.

The cluster can be started with this command:
```shell
$ docker-compose up -d
```

After this command return, run `docker-compose ps` to check whether all services are started correctly.
If no error occurs, the gateway proxy service should be listening at `0.0.0.0:8080`,
while admin API is serving on `0.0.0.0:8082`.

### Standalone

As described above, the apron gateway service requires redis for saving service and user key data,
In previous docker-compose mode, the redis service will be started in the docker-compose cluster.
While for the standalone setup, the service should be set before starting the gateway.
The instruction below will use `localhost:6379` as the redis server address.

This demo presents request forward to local httpbin service, which is running in a docker instance.
And, the public httpbin service (http://httpbin.org) can also be used.

```shell
$ docker run -it --rm -p 2345:80 kennethreitz/httpbin
```

The binary name built by `make build` command is *gw*, which should be at top directory of the project.
The configuration can be modified with environment variable, which are:

* PROXY_PORT: listening port for proxy service, should be int value between 1 and 65535.
* ADMIN_ADDR: listening address for admin service, should be a full address such as *0.0.0.0:8082*
* REDIS_SERVER: redis service address, should be in the format of *<IP>:<PORT>*, such as *localhost:6379*

The service can be started with this command, if the environment variables listed above not set,
the default value will be used.

```
./gw
```

## Usage demo

> In this section, the admin API address is *http://localhost:8082*
> while the proxy address is *http://localhost:8080*

### Create a service

*POST /service/*

| Param    | Type   | Desc                                                         | Sample value           |
| -------- | ------ | ------------------------------------------------------------ | ---------------------- |
| name     | string | Name of service, will be used while generating key           | `test_httpbin_service` |
| base_url | string | Base url or name for service, all request will be forwarded to this | httpbin/               |
| schema   | string | Schema for building service, support http, https, ws, wss    | http                   |



```shell
$ http -j post http://localhost:8082/service/ name=test_httpbin_service base_url=httpbin/ schema=http

```

If service created successfully, service will return status 201.

### Create a user key

*POST /service/<service_name>/keys/*

| Params     | Type   | Desc                   | Sample value    |
| ---------- | ------ | ---------------------- | --------------- |
| account_id | string | Account id of this key | test_account_id |



```shell
$ http post http://localhost:8082/service/test_httpbin_service/keys/ account_id=test_account_id
```

This is a sample response

```json
{
    "issuedAt": "1615338539",
    "key": "a6d9c1b2-0bc0-43ec-b8b1-f4aa5a443288",
    "serviceName": "test_httpbin_service",
  	"accountId": "test_account_id"
}
```

### Access the service via proxy

*GET /v1/<service_name>/<user_key>/<requests>*

```shell
$ http http://localhost:8080/v1/test_httpbin_service/a6d9c1b2-0bc0-43ec-b8b1-f4aa5a443288/anything/foobar
```

This is a sample reponse

```json
{
    "args": {},
    "data": "",
    "files": {},
    "form": {},
    "headers": {
        "Accept": "*/*",
        "Accept-Encoding": "gzip, deflate",
        "Connection": "keep-alive",
        "Content-Length": "0",
        "Host": "httpbin",
        "User-Agent": "HTTPie/2.2.0"
    },
    "json": null,
    "method": "GET",
    "origin": "172.23.0.3",
    "url": "http://httpbin/anything/foobar"
}
```

### Get usage report

*GET /service/report/*

```shell
$ http http://localhost:8082/service/report/
```

This is a sample response

```json
[
    {
        "Cost": 0,
        "end_time": 1615338703,
        "id": 1615338595,
        "price_plan": "",
        "service_uuid": "test_httpbin_service",
        "start_time": 1615338595,
        "usage": 2,
        "user_key": "a6d9c1b2-0bc0-43ec-b8b1-f4aa5a443288"
    }
]
```
