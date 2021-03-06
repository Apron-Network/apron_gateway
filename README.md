# The Gateway for Apron Project

## Build
```
make
```

## Testing

Launch demo backend service
```
docker run -it --rm -p 2345:80 kennethreitz/httpbin
```

or

```
docker-compose up -d
```

Launch gateway
```
./gw
```

Test proxy
```
http get localhost:8080
```

Test admin
```
http get localhost:8081
```