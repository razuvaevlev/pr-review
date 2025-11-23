# Тестовое задание Авито осень 2025

## Запуск

```bash
cp env.example .env
docker-compose up --build
```
## Цели make
```bash
Usage: make [target]

Targets:
  build       Build the binary
  run         Build and run the service
  test        Run unit tests
  fmt         Run gofmt (in-place)
  vet         Run go vet
  lint        Run golangci-lint (if installed)
  deps        Download Go modules
  clean       Remove generated binaries
```
