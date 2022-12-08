FROM golang:alpine AS builder
WORKDIR /app
COPY ./ ./
RUN go build -o main ./cmd/

FROM alpine:latest
COPY --from=builder /app/main /main
COPY --from=builder /app/tasks.csv /tasks.csv
COPY --from=builder /app/options.yml /options.yml
COPY --from=builder /app/token.yml /token.yml
COPY --from=builder /app/pics/ /pics/
COPY --from=builder /app/Глава_1.doc /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/main"]