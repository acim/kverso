FROM golang:1.12-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-s -w" -o kverso

FROM alpine

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/kverso /usr/bin/kverso

ENTRYPOINT ["/usr/bin/kverso"]