FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP_CMD=api
RUN go build -o /bin/banking-service ./cmd/${APP_CMD}

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /bin/banking-service /bin/banking-service

EXPOSE 8080

ENTRYPOINT ["/bin/banking-service"]
