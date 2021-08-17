This bot drops a TikTok preview into Discord channels whenever
somebody mentions a share URL.

# Build & run

```
$ go build ./cmd/dustybot
$ DISCORD_TOKEN="NNNNNNN.NNNNN.NNNNNN" ./dustybot
```

# Or in Docker

```
$ docker build -t dustybot .
$ docker run -it --rm dustybot
```