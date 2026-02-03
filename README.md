# Notifications Service

A Go-based notification platform that delivers messages via **email**, **push** (console), and **WebSocket**. It uses Kafka for async message delivery, Redis for WebSocket session state, and MailHog for local email testing.

## Service Overview

- **async-notifications** — Consumes notification events from Kafka and processes them: sends emails via SMTP (MailHog) and logs push notifications to the console. Failed messages are retried and, after max retries, sent to a dead-letter topic.
- **ws-notifications** — Exposes an HTTP server with a WebSocket endpoint. Clients subscribe by email; notifications for that user are consumed from Kafka, written to Redis, and pushed to connected WebSocket clients in real time. Supports reconnection and offline message handling.

**Stack:** Go, Kafka, Zookeeper, Redis, Redis Insight, MailHog, Kafka UI.

---

## Quick Start

### Build and run all services (Docker)

Build images and start the app plus infrastructure (Kafka, Zookeeper, Redis, MailHog, Kafka UI, Redis Insight):

```bash
docker compose build && docker compose up
```

**Description:** Builds the Go apps (async-notifications and ws-notifications) and starts every service defined in `docker-compose.yml`. The WebSocket API is available on port 8080.

---

## Development

### Run tests (with race detector and sync test experiment)

Clear the terminal, then run all tests using the `synctest` experiment (for sync-based test helpers):

```bash
clear && GOEXPERIMENT=synctest go test ./...
```

**Description:** Clears the screen and runs the full test suite. `GOEXPERIMENT=synctest` enables Go’s sync test experiment. Use this for local verification before pushing.

### Generate mocks for HTTP layer

Generate mock implementations of the service interfaces used by the HTTP controller:

```bash
mockgen -source=internal/service/contracts.go -destination=internal/service/mocks.go -package=http
```

**Description:** Uses `mockgen` to create mocks. For example, from `internal/service/contracts.go` and write them to `internal/service/mocks.go` in package `http`. Run this after changing interfaces and so controller tests keep compiling.

### Lint (optional alias)

Use a fixed golangci-lint binary so CI and local checks stay aligned:

```bash
alias glint='/opt/homebrew/Cellar/golangci-lint/2.8.0/bin/golangci-lint'
```

**Description:** Defines a shell alias for the specific golangci-lint version. After this, run `glint run` (or `glint run ./...`) to lint the project. Adjust the path if your Homebrew install location differs.

---

## API

### WebSocket subscribe

Subscribe to real-time notifications for a user by email. The connection is upgraded to WebSocket.

**Endpoint:** `GET /notifications/subscribe?userEmail=<email>`

**Example:**

```
http://localhost:8080/notifications/subscribe?userEmail=w1@rty.ru
```

**Description:** Open this URL in a WebSocket-capable client (browser or tool). The server upgrades the request to WebSocket and associates the connection with the given `userEmail`. All notifications for that user (from Kafka → Redis) are then pushed over this connection.

### Sending a notification (payload example)

Notifications are typically produced into Kafka by other services. The payload shape for a notification (e.g. for email or WS) looks like:

```json
{
  "user_email": "w1@rty.ru",
  "subject": "Hello",
  "body": "я пришел к тебе с приветом"
}
```

**Description:** `user_email` identifies the recipient; `subject` and `body` are used for email and can be used for other channels. Produce messages in this format to the appropriate Kafka topics (e.g. email, push, or WS) so that async-notifications or ws-notifications can consume and process them.

---

## Infrastructure URLs (when using Docker Compose)

When running `docker compose up`, these UIs and ports are available:

| Service      | URL                     | Description                                      |
|-------------|-------------------------|--------------------------------------------------|
| Kafka UI    | http://localhost:8000   | Manage topics, consumer groups, and messages     |
| MailHog     | http://localhost:8025/  | Inbox for emails sent via SMTP (port 1025) login: test, password: test      |
| Redis Insight | http://localhost:5540/ | Inspect Redis data used by ws-notifications      |

**Note:** Kafka is on port 9092, Redis on 6379, MailHog SMTP on 1025. Use these for configuring producers or debugging.

---

## Test scenarios (WebSocket)

Useful cases to verify WebSocket and offline behavior:

- **Client disconnect** — Client closes the connection; server should clean up and stop sending to that connection.
- **Server closes connection** — Server closes the WebSocket; client should detect closure and can reconnect.
- **Message while user offline** — Send a notification for a user with no active WebSocket; it should be stored (e.g. in Redis) and delivered on next connection/reconnect if the design supports it.
- **Reconnect with same email** — Client disconnects and reconnects with the same `userEmail`; server should accept the new connection and associate it with that user again.

---

## Summary of commands

| Command | Description |
|--------|-------------|
| `docker compose build && docker compose up` | Build and run the whole stack (apps + Kafka, Redis, MailHog, etc.). |
| `clear && GOEXPERIMENT=synctest go test ./...` | Run full test suite with sync experiment and a clear screen. |
| `mockgen -source=internal/service/contracts.go -destination=internal/service/mocks.go -package=http` | Regenerate HTTP-layer mocks from service contracts. |
| `alias glint='/opt/homebrew/Cellar/golangci-lint/2.8.0/bin/golangci-lint'` | Alias for a specific golangci-lint binary (path may vary). |
