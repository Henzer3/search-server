# XKCD Search Service

Search and indexing service for XKCD comics with database-backed search, in-memory index search, JWT-protected critical operations, event-driven index synchronization through NATS, and metrics exposure for monitoring.

# Recent Additions
- Event-driven index rebuild via NATS instead of frequent polling
- Fast in-memory search endpoint alongside database search
- JWT login and authorization middleware for critical operations
- Concurrency limiter for database-backed search
- Rate limiter for indexed search
- Request duration metrics with VictoriaMetrics-compatible /metrics endpoint
- Grafana/VictoriaMetrics monitoring integration
- Docker Compose based local environment

The goal of this project is to demonstrate a production-oriented Go service structure with clear separation between core logic, adapters, and middleware, while keeping the system easy to understand and extend.

# Technology stack:
- Go
- PostgreSQL
- NATS
- VictoriaMetrics
- Grafana
- Docker / Docker Compose
- Rest api
- grpc


Features
- /api/search — search through PostgreSQL
- /api/isearch — search through in-memory index
- /api/login — JWT token выдача для суперпользователя
- protected critical endpoints for database update and reset
- event-driven index rebuild on database update
- index reset on database cleanup
- request limiting and monitoring
- Architecture

The project is split into several services/components:

- update — updates the XKCD database and publishes events
- search — performs search and maintains the in-memory index
- words — text processing / normalization support
- NATS — event bus between services
- PostgreSQL — source of truth
- VictoriaMetrics + Grafana — metrics storage and visualization
