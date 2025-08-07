# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o nut_exporter .

# Minimal runtime image
FROM alpine:3.20

COPY --from=builder /app/nut_exporter /usr/local/bin/nut_exporter

EXPOSE 8100
ENTRYPOINT ["/usr/local/bin/nut_exporter"]
