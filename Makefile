.PHONY: all db
all: db

db:
	podman stop gophermart || true
	podman rm gophermart || true
	podman run --name gophermart -p 5432:5432 -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -d postgres:16.1
	until podman exec gophermart pg_isready -h localhost -p 5432; do \
  		echo "waiting Postgres to be ready..."; \
  		sleep 1; \
  	done
	podman exec -e PGPASSWORD=${POSTGRES_PASSWORD} gophermart psql -U postgres -c "CREATE DATABASE gophermart;"
