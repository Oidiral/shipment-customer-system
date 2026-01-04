# Shipment-Customer System

Микросервисная система для управления заявками на отгрузку с REST API, gRPC коммуникацией и распределённым трейсингом.

## Как запустить

```bash
docker-compose up
```

Миграции применятся автоматически.

## Пример REST

### Создать заявку

```bash
curl -X POST http://localhost:8080/api/v1/shipments \
  -H "Content-Type: application/json" \
  -d '{"route":"ALMATY→ASTANA","price":120000,"customer":{"idn":"990101123456"}}'
```

Ответ (201):
```json
{
  "id": "uuid",
  "status": "CREATED",
  "customerId": "uuid"
}
```

### Получить заявку

```bash
curl http://localhost:8080/api/v1/shipments/<id>
```

Ответ (200):
```json
{
  "id": "uuid",
  "route": "ALMATY→ASTANA",
  "price": 120000,
  "status": "CREATED",
  "customerId": "uuid",
  "created_at": "2025-01-04T12:00:00Z"
}
```

## Где смотреть трассы

Jaeger UI: **http://localhost:16686**

В Jaeger виден trace цепочки: **REST → shipment-service → gRPC → customer-service → DB**

Выберите сервис `shipment-service`, и вы увидите полную цепочку вызовов от входящего HTTP-запроса до операций с базой данных.

## Как это работает

REST-запрос приходит на Envoy (порт 8080), который проксирует его в shipment-service. Shipment-service через Envoy (внутренний порт 9090) делает gRPC-вызов к customer-service для upsert клиента. gRPC-порт доступен только внутри docker-сети и не пробрасывается наружу.

## Архитектура

```
[Client] → REST (Envoy:8080) → shipment-service → gRPC (Envoy:9090) → customer-service → Postgres
```

## Структура проекта

```
/cmd/shipment-service/main.go   # точка входа shipment-service
/cmd/customer-service/main.go   # точка входа customer-service
/api/proto/customer.proto       # gRPC протокол
/internal/shipment/             # shipment-service логика
/internal/customer/             # customer-service логика
/internal/telemetry/            # OpenTelemetry
/config/envoy.yaml              # конфигурация Envoy
/config/otel-collector.yaml     # конфигурация OTel Collector
/migrations/                    # SQL миграции
/docker-compose.yaml            # оркестрация
```

## Технологии

- **REST API**: HTTP через Envoy (внешний доступ)
- **gRPC**: внутренняя коммуникация между сервисами (через Envoy, не доступно извне)
- **PostgreSQL**: база данных с двумя таблицами (customers, shipments)
- **OpenTelemetry**: распределённый трейсинг
- **Jaeger**: визуализация трейсов
- **Envoy**: API Gateway с rate limiting 