# Subscriptions Service

REST service for aggregating user subscription data.

## Tech stack

- Go
- Echo
- PostgreSQL
- pgx
- docker-compose
- golang-migrate
- Swagger (to be added)

## Run project

Start database:

docker compose up -d


Run migrations:


docker compose up migrate


Run service:


go run ./cmd/app


Health check:


curl http://localhost:8080/health


## Create subscription


curl -X POST http://localhost:8080/api/v1/subscriptions

-H "Content-Type: application/json"
-d '{
"service_name":"Yandex Plus",
"price":400,
"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba",
"start_date":"07-2025"
}'