FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/server
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY migrations ./migrations
USER app
EXPOSE 8080
ENTRYPOINT ["api"]
