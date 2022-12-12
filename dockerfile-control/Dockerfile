FROM golang:alpine AS builder
WORKDIR /app
COPY .. ./
RUN go build -o main ./cmd/control_bot/

FROM alpine:latest
COPY --from=builder /app/main /main

ENTRYPOINT ["/main"]