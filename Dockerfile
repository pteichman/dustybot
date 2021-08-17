FROM golang:alpine AS builder

WORKDIR $GOPATH/src
COPY . .

RUN go build ./cmd/dustybot -o /go/bin/dustybot

FROM scratch
COPY --from=builder /go/bin/dustybot /go/bin/dustybot

ENTRYPOINT ["/go/bin/dustybot"]