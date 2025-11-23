FROM golang:1.25.4

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["/bin/sh", "-c", "until pg_isready -h ${DB_HOST} -U ${DB_USER} -d ${DB_TEST_NAME}; do sleep 1; done; go test ./test/... -v && touch /tmp/tests_ok && sleep infinity || (echo 'tests failed' && exit 1)"]
