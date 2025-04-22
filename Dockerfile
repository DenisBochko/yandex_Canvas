# Stage 1: Builder
FROM golang:1.23.1 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/bin/canvas ./cmd/canvas/main.go

# Stage 2: Run 
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/canvas /app/bin/canvas
COPY --from=builder /app/db /app/db

EXPOSE 50051

CMD ["/app/bin/canvas"]