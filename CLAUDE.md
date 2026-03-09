# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**paas-finops-extend** is a Go + Gin FinOps API backend. It runs two servers concurrently:
- **Port 8888** – Main HTTP REST API (admin management + alerting)
- **Port 8443** – HTTPS Kubernetes mutating admission webhook (pod CPU/memory mutation)

Requires MySQL 5.7+ with schema from `static-files/finops_extend_schema.sql`.

## Build and Run

```bash
go build -o server main.go
./server

# Optional flags
./server -c /path/to/config.yaml    # override config file

# Config loading priority: -c flag > GVA_CONFIG env var > ./config.yaml

go mod tidy

docker build -t paas-finops-extend .
docker run -p 8888:8888 -p 8443:8443 paas-finops-extend
```

## Architecture

**Request Flow:**
```
Routes (router/) → API Handlers (api/v1/) → Services (service/) → Models (model/) → GORM → MySQL
```

**Module Structure:**
Four modules follow the same layered pattern: `manage`, `observe`, `webhook`, and `example` (template only, no routes).

```
api/v1/{module}/      → HTTP handlers
service/{module}/     → Business logic
model/{module}/       → GORM models + request/response structs
router/{module}/      → Route registration
```

**Entry Points (`enter.go` files):**
Each layer aggregates sub-groups via a top-level `enter.go`:
- `service/enter.go` → `ServiceGroupApp` (Manage, Observe, Example, Webhook)
- `router/enter.go` → `RouterGroupApp` (Manage, Observe, Webhook)
- `api/v1/enter.go` → `ApiGroupApp` (Manage, Observe, Webhook)

**Global Variables (`global/global.go`):**
- `GVA_DB` – GORM database instance
- `GVA_VP` – Viper config handler
- `GVA_LOG` – Zap logger
- `GVA_CONFIG` – Parsed config (`config.Server`)
- `GVA_K8S_DYNAMIC` – K8s dynamic client (`dynamic.Interface`)
- `GVA_K8S_INDEXER` – K8s cache indexer for Recommendation CRs

## API Routes

**Management (`/api/v1/manage/`):**
- `POST adminUser/login` – Login (public)
- `POST createadminUser` – Create user (auth required)
- `PUT adminUser/name` / `PUT adminUser/password` – Update profile
- `GET adminUser/profile` – Get profile
- `DELETE logout` – Logout
- `POST upload/file` – File upload

**Alerts (`/api/v1/observe/`):**
- `POST/GET alerts` – Create / list alerts
- `GET/PUT/DELETE alerts/:alertId` – Get / update / delete by ID
- `DELETE alerts` – Batch delete

**Other:**
- `GET /health` – Health check
- `POST /mutate` (port 8443, HTTPS) – K8s admission webhook

## Key Patterns

**Standard Response:**
```go
response.OkWithData(data, c)        // resultCode: 200
response.OkWithMessage("msg", c)    // resultCode: 200
response.FailWithMessage("err", c)  // resultCode: 500
// Unauthenticated responses use resultCode: 416
```

**Service Injection:**
```go
var alertService = service.ServiceGroupApp.ObserveServiceGroup.ObserveAlertService
```

**Model Definition:**
```go
type PrometheusAlert struct {
    AlertId int `json:"alertId" gorm:"primarykey;AUTO_INCREMENT"`
    // ...
}
func (PrometheusAlert) TableName() string { return "prometheus_alert" }
```

**JWT Auth:** Uses custom `token` header (not `Authorization`). Token is validated against the `admin_user_token` table.

**Alert Deduplication:** Fingerprint is generated from alert labels and stored in `prometheus_alert.fingerprint` (unique index). Upsert pattern tracks `alert_count`, `daily_notify_count`, and `last_notify_date`. Daily notification limit is enforced per the `mq.daily-notify-limit` config.

**Adding a New Module:**
1. Create `api/v1/{module}/enter.go` with API group struct
2. Create `service/{module}/enter.go` with service group struct
3. Create `model/{module}/` with GORM models
4. Create `router/{module}/enter.go` with router group struct
5. Register in all parent `enter.go` files and `initialize/router.go`

## Kubernetes Webhook

The webhook is a mutating admission webhook that patches pod CPU/memory requests based on `Recommendation` CRs (group: `bcs.finops.io`, version: `v1alpha1`).

**Flow:** `POST /mutate` → `RecommendationApi.ServeMutate()` → `RecommendationService.MutatePod()` → JSON Patch response

**Indexer key:** `{clusterId}/{namespace}/{workloadName}` — enables O(1) cache lookup.

**Validation:** Recommended CPU is capped at the container's existing limit to prevent pod startup failure.

TLS cert/key paths are configured via `system.tls-cert` / `system.tls-key` in `config.yaml`.

## Middleware

- `middleware.AdminJWTAuth()` – Token validation for protected admin routes
- `middleware.Cors()` – CORS handling (supports POST, GET, OPTIONS, DELETE, PUT)

## Tech Stack

- **Go:** 1.24+ | **Framework:** Gin v1.11 | **ORM:** GORM v1.31
- **Config:** Viper (`config.yaml`) | **Logging:** Uber Zap (logs to `/log/`)
- **Database:** MySQL 5.7+ | **K8s:** `k8s.io/client-go` dynamic client

## Database Tables

- `admin_user` – Users; default: `admin` / `123456` (MD5 hashed)
- `admin_user_token` – JWT tokens keyed by `admin_user_id`
- `prometheus_alert` – Alerts with dedup fingerprint and notification tracking
