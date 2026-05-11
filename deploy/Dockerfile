FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY . .

RUN go build -o /bin/banking-service ./cmd/api

EXPOSE 8080

CMD ["/bin/banking-service"]
