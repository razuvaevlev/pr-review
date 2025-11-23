FROM golang:1.25.4 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -v -o /out/pr-review ./cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl postgresql-client

COPY --from=builder /out/pr-review /usr/local/bin/pr-review
COPY --from=builder /src/migrations /migrations
COPY entrypoint.sh /usr/local/bin/entrypoint.sh

EXPOSE 8080
ENV PORT=8080

RUN chmod +x /usr/local/bin/entrypoint.sh

USER nobody

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]