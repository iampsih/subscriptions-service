# Subscriptions Service (Effective Mobile test task)

REST service for aggregating user online subscriptions.

## Features

- CRUDL for subscriptions
- Total cost calculation for a selected period (month granularity)
- PostgreSQL + migrations
- Structured logs + request id
- Swagger UI
- Run with docker compose

## Tech stack

- Go
- Echo (v4)
- PostgreSQL
- pgxpool
- golang-migrate (migrate/migrate)
- swaggo (Swagger)

## Data model

Subscription:
- `service_name` (string)
- `price` (int, RUB/month)
- `user_id` (UUID)
- `start_date` (`MM-YYYY`)
- `end_date` (`MM-YYYY`, optional)

Month values are stored in DB as `DATE` with the first day of month (`YYYY-MM-01`).

## Run (docker)

Create `.env` (you can copy from `.env.example`):
```bash
cp .env.example .env

Start:

docker compose up --build
Endpoints

Base path: /api/v1

POST /subscriptions

GET /subscriptions/{id}

PUT /subscriptions/{id}

DELETE /subscriptions/{id}

GET /subscriptions (list, filters + pagination)

GET /subscriptions/total (sum for period)

Health:

GET /health

Swagger:

GET /swagger/index.html