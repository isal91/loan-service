#!/bin/bash
# scripts/setup.sh

echo "🚀 Starting Local Environment..."

# 1. Start Environment (Postgres & Swagger UI)
docker-compose up -d

# 2. Wait for Postgres to be ready
echo "⏳ Waiting for database..."
until docker exec loan-service-db pg_isready; do
  sleep 1
done

# 3. Setup .env if it doesn't exist
if [ ! -f .env ]; then
  if [ -f .env.example ]; then
    cp .env.example .env
    echo "✅ Created .env from .env.example"
  else
    echo "⚠️ .env.example not found, skipping .env creation"
  fi
fi

# 4. Run Migrations
echo "📂 Running database migrations..."
for f in migrations/*.sql; do
    echo "  -> Applying $f"
    docker exec -i loan-service-db psql -U user -d loan_db < "$f"
done

# 5. Run Go application
echo "🏃 Starting Loan Service..."
# Kill existing process on port 8080 if any
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
go run cmd/main.go
