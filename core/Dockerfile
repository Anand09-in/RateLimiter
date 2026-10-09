FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /rate-limiter-core ./cmd/server

FROM builder AS test
RUN apk --no-cache add gcc musl-dev
# REDIS_ADDR is injected at runtime via --env
CMD ["go", "test", "-race", "-v", "-count=1", "./internal/algorithms/..."]

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /rate-limiter-core .

EXPOSE 8080
ENTRYPOINT ["./rate-limiter-core"]
