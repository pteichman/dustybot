FROM golang:alpine AS builder

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -o /dustybot ./cmd/dustybot

FROM scratch
COPY --from=builder /dustybot /dustybot
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/dustybot"]