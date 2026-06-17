# Webhook Delivery Engine

A production-grade webhook delivery system built in Go. Accepts event ingestion, fans out to registered subscribers, and delivers HTTP notifications reliably with automatic retries, HMAC signing, and a full delivery audit log.

Inspired by the infrastructure behind Stripe, GitHub, and Svix webhooks.

---

## Architecture

```
POST /events
     │
     ▼
┌─────────────────────────────────┐
│  Ingest Handler                 │
│  BEGIN TRANSACTION              │
│  → insert into events           │
│  → find matching webhooks       │
│  → insert into outbox (fan-out) │
│  COMMIT                         │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  Outbox Poller (every 1s)       │
│  → fetch pending outbox rows    │
│  → enqueue asynq jobs           │
│  → mark rows as enqueued        │
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  Worker (asynq + Redis)         │
│  → sign payload with HMAC-SHA256│
│  → HTTP POST to target URL      │
│  → log attempt to delivery_logs │
│  → retry on failure (backoff)   │
│  → mark outbox as delivered     │
└─────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go |
| HTTP Router | chi |
| Database | PostgreSQL (pgx/v5) |
| Query Generation | sqlc |
| Migrations | goose |
| Job Queue | asynq + Redis |
| Signing | HMAC-SHA256 |

---

## Key Engineering Decisions

**Transactional Outbox Pattern**
Event ingestion and outbox row creation happen in a single database transaction. This guarantees no event is ever lost even if the process crashes between write and delivery — the outbox row always exists before the process moves on.

**asynq for reliable job processing**
asynq provides at-least-once delivery semantics with configurable retry backoff, concurrency control, and a dead letter queue. Failed deliveries are automatically retried with exponential backoff without any custom retry logic.

**HMAC-SHA256 payload signing**
Every outbound webhook request includes an `X-Webhook-Signature: sha256=...` header. Subscribers can verify the signature using their shared secret to confirm the request is genuine and untampered — the same approach used by Stripe and GitHub.

**Fan-out per subscriber**
One incoming event creates one independent outbox row per matching webhook. Failures are isolated — a failing subscriber doesn't affect delivery to others, and each is retried independently.

---

## Running Locally

**Prerequisites:** Go 1.21+, Docker Desktop

```bash
# 1. clone the repo
git clone https://github.com/heythisissud/webhook-engine
cd webhook-engine

# 2. start postgres and redis
docker-compose up -d

# 3. run migrations
goose -dir migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=webhookdb sslmode=disable" up

# 4. create .env
cp .env.example .env

# 5. start the server
go run cmd/server/main.go
```

---

## Environment Variables

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/webhookdb
REDIS_ADDR=localhost:6379
PORT=8080
```

---

## API Reference

### Register a webhook
```bash
curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "user_1",
    "target_url": "https://your-server.com/webhook",
    "secret": "your-secret",
    "event_types": ["payment.success", "payment.failed"]
  }'
```

### List webhooks
```bash
curl http://localhost:8080/webhooks?client_id=user_1
```

### Get a webhook
```bash
curl http://localhost:8080/webhooks/{id}
```

### Delete a webhook
```bash
curl -X DELETE http://localhost:8080/webhooks/{id}
```

### Ingest an event
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "user_1",
    "event_type": "payment.success",
    "payload": {"amount": 100, "currency": "usd"}
  }'
```

Returns `202 Accepted` immediately. Delivery happens asynchronously.

---

## Database Schema

```
webhooks        — subscriber registrations
events          — raw incoming events
outbox          — delivery jobs (one per webhook per event)
delivery_logs   — full audit log of every delivery attempt
```

---

## How Delivery Works

1. Event arrives at `POST /events`
2. Saved to DB + outbox rows created atomically in one transaction
3. Poller picks up pending rows every second, pushes to Redis queue
4. Worker dequeues job, signs payload, makes HTTP POST
5. On success → outbox marked `delivered`, attempt logged
6. On failure → asynq retries automatically with backoff
7. After max retries → job moved to dead letter queue

---

## Project Structure

```
cmd/server/         — entrypoint
internal/
  api/              — HTTP handlers
  config/           — environment config
  db/
    queries/        — SQL query files
    generated/      — sqlc generated code
  delivery/         — HTTP delivery logic
  models/           — request/response structs
  poller/           — outbox poller goroutine
  worker/           — asynq task definitions and handlers
migrations/         — goose SQL migrations
docker-compose.yml
```
