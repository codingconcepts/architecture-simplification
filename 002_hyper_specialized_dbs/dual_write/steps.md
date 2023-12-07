# Before

### Infra

``` sh
(
  cd 002_hyper_specialized_dbs/dual_write/before && \
  docker compose up --build --force-recreate -d
)
```

### Run

``` sh
go run 002_hyper_specialized_dbs/dual_write/before/main.go
```

### Teardown

``` sh
make teardown
```