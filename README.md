# certificatemgmt

HTTP service backed by MongoDB; configuration is supplied via environment variables.

## Run with Docker

Build the image from the repository root:

```bash
docker build -t certificatemgmt:latest .
```

Start a container. To point at MongoDB running on the host machine, use `host.docker.internal` on Docker Desktop for macOS or Windows:

```bash
docker run --rm -p 8080:8080 \
  -e MONGODB_URI='mongodb://host.docker.internal:27017' \
  -e MONGODB_DATABASE=app \
  certificatemgmt:latest
```

The service listens on **8080** by default. Main environment variables:

| Variable | Description |
|----------|-------------|
| `MONGODB_URI` | Required. MongoDB connection string |
| `MONGODB_DATABASE` | Optional. Database name; default `app` |
| `PORT` | Optional; default `8080` |
| `HTTP_ADDR` | Optional. If set, overrides the listen address from `PORT` (for example `:3000`) |

## API

- `GET /healthz` — Liveness probe
- `GET /readyz` — Readiness probe (checks MongoDB)
- `POST /v1/assign` — Body `{"uid":"<user id>"}`; assigns an account (when possible) and returns account details

## Run locally (without Docker)

```bash
export MONGODB_URI='mongodb://localhost:27017'
go run .
```
