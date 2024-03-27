# Polite Rate Limit

In-memory token-bucket rate limiter wrapped around a small Go HTTP API.

- `/hello` — limited (5 tokens/sec, burst 10) per client IP
- `/health` — unlimited liveness probe

## Run

```bash
go test ./...
go run .
curl -i http://127.0.0.1:8080/hello
```

Hammer `/hello` to see `429` with `Retry-After`.

## Caveats

In-memory buckets are per-process. For multiple replicas, back the bucket with Redis (or an edge limiter).

## License

MIT
