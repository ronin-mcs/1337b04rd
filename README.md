# 1337b04rd

1337b04rd is a small anonymous imageboard-style web application written in Go.
It stores posts, comments, sessions, anonymous user metadata, and attachment
records in PostgreSQL. Uploaded files are stored in the bundled `triple-s`
service, a simple S3-like object storage implementation included in this
repository.

## Features

- Create posts with text and optional file attachments.
- Add comments and replies to existing posts.
- Show post catalog and archive pages with attachment previews.
- Assign anonymous users Rick and Morty avatar IDs.
- Store uploaded files in `triple-s`.
- Persist application data in PostgreSQL.

## Tech Stack

- Go 1.22
- PostgreSQL 15
- Docker Compose
- HTML templates from the `templates/` directory
- Custom `triple-s` object storage service

## Project Structure

```text
cmd/                  Application entrypoint and logging setup
internal/adapters/    HTTP, PostgreSQL, S3, and external API adapters
internal/domain/      Core board use cases and service logic
models/               Shared domain models
templates/            HTML templates
triple-s/             Local S3-like storage service
init.sql              PostgreSQL initialization script
docker-compose.yml    Local development stack
```

## Running with Docker Compose

Start the full stack:

```bash
docker compose up --build
```

Run it in the background:

```bash
docker compose up -d --build
```

The application will be available at:

```text
http://localhost:8080/posts/create
```

The `triple-s` service is exposed at:

```text
http://localhost:9000
```

PostgreSQL is exposed at:

```text
localhost:5432
```

## Services

The Compose stack starts three services:

- `app`: the Go web application.
- `postgres`: PostgreSQL database initialized from `init.sql`.
- `triple-s`: local object storage for uploaded files.

## Environment Variables

The app container uses these variables:

```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=app

S3_ENDPOINT=http://triple-s:9000
S3_PUBLIC_ENDPOINT=http://localhost:9000
S3_BUCKET=1337b04rd
```

`S3_ENDPOINT` is used by the app container to upload and delete files inside the
Docker network.

`S3_PUBLIC_ENDPOINT` is used when rendering image links in HTML. It must be
reachable from the browser, so for local Docker runs it should usually be
`http://localhost:9000`.

## Routes

Main browser routes:

```text
GET  /posts
GET  /posts/create
POST /posts/create
GET  /posts/{id}
POST /posts/{id}/comments
POST /posts/{id}/comments/{comment_id}/replies
GET  /archive
GET  /archive/{id}
```

## Database Initialization

PostgreSQL runs `init.sql` only when the database volume is created for the
first time. If the `postgres-data` volume already exists, changes to `init.sql`
will not be applied automatically.

To recreate the database from scratch:

```bash
docker compose down -v
docker compose up --build
```

Warning: `docker compose down -v` removes Compose volumes, including database
data and stored `triple-s` files.

## Rebuilding After Code Changes

The app image copies the source code and compiles a Go binary during the Docker
build:

```dockerfile
COPY . .
RUN go build -o app ./cmd/.
```

Because of that, `docker start 1337b04rd` will not pick up code changes. Rebuild
and recreate the app container instead:

```bash
docker compose up -d --no-deps --build --force-recreate app
```

For a full no-cache rebuild:

```bash
docker compose build --no-cache
docker compose up -d
```

## Running Tests

Run all Go tests:

```bash
go test ./...
```

If your Go build cache directory is not writable, use a temporary cache:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
```

## Troubleshooting

If the app logs `connect: connection refused` for PostgreSQL, the app started
before the database was ready. Restart the app after PostgreSQL is healthy:

```bash
docker compose up -d postgres
docker compose logs -f postgres
docker compose up -d --no-deps --build --force-recreate app
```

If uploaded images do not load in the browser, check the rendered `<img src>`.
It should use `http://localhost:9000/...`, not `http://triple-s:9000/...`.

If tables are missing, recreate the database volume or apply `init.sql`
manually. The init script is only executed on the first PostgreSQL volume
initialization.
