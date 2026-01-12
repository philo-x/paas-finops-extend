# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **paas-finops-extend**, a Go + Gin implementation of a FinOps API backend with admin user management and observability (alerts) functionality.

## Build and Run Commands

```bash
# Build
go build -o server main.go

# Run
./server

# Dependencies
go mod tidy -compat=1.17
```

The server runs on port 8888 (configurable in `config.yaml`). Requires MySQL 5.7+ with schema from `/static-files/finops_extend_schema.sql`.

## Architecture

**Request Flow:**
```
Routes (router/) → API Handlers (api/v1/) → Services (service/) → Models (model/) → GORM → MySQL
```

**Module Structure:**
The codebase uses a group-based architecture with three modules: `manage`, `observe`, and `example`. Each module follows the same layered pattern:

```
api/v1/{module}/      → HTTP handlers
service/{module}/     → Business logic
model/{module}/       → GORM models + request/response structs
router/{module}/      → Route registration
```

**Entry Points (enter.go files):**
Each layer has an `enter.go` that aggregates module groups:
- `service/enter.go` → `ServiceGroupApp` (contains ManageServiceGroup, ObserveServiceGroup, ExampleServiceGroup)
- `router/enter.go` → `RouterGroupApp` (contains Manage, Observe router groups)
- `api/v1/enter.go` → `ApiGroupApp` (contains ManageApiGroup, ObserveApiGroup)

**Global Variables:**
- `global.GVA_DB` - GORM database instance
- `global.GVA_LOG` - Zap logger
- `global.GVA_CONFIG` - Viper configuration

## API Routes

**Base paths:**
- `/api/manage/v1/` - Admin management routes
- `/api/observe/v1/` - Observability routes (alerts)

**Admin User (`/api/manage/v1/`):**
- `POST adminUser/login` - Admin login (public)
- `POST createadminUser` - Create admin user (auth required)
- `PUT adminUser/name` - Update admin name
- `PUT adminUser/password` - Update password
- `GET adminUser/profile` - Get admin profile
- `DELETE logout` - Logout
- `POST upload/file` - Upload file

**Alerts (`/api/observe/v1/`):**
- `POST alerts` - Create alert
- `GET alerts` - Get alert list
- `GET alerts/:alertId` - Get alert by ID
- `PUT alerts/:alertId` - Update alert
- `DELETE alerts/:alertId` - Delete alert
- `DELETE alerts` - Batch delete alerts

## Key Patterns

**Standard Response:**
```go
response.OkWithData(data, c)
response.OkWithMessage("message", c)
response.FailWithMessage("error", c)
```

**Service Injection:**
```go
var finopsAdminUserService = service.ServiceGroupApp.ManageServiceGroup.ManageAdminUserService
var alertService = service.ServiceGroupApp.ObserveServiceGroup.ObserveAlertService
```

**Model Definition:**
```go
type AdminUser struct {
    AdminUserId int `json:"adminUserId" gorm:"primarykey;AUTO_INCREMENT"`
    // ...
}
func (AdminUser) TableName() string {
    return "admin_user"
}
```

**Adding a New Module:**
1. Create `api/v1/{module}/enter.go` with API group struct
2. Create `service/{module}/enter.go` with service group struct
3. Create `model/{module}/` with GORM models
4. Create `router/{module}/enter.go` with router group struct
5. Register in parent `enter.go` files and `initialize/router.go`

## Middleware

- `middleware.AdminJWTAuth()` - JWT authentication for admin routes
- `middleware.Cors()` - CORS handling

## Tech Stack

- **Go:** 1.17
- **Framework:** Gin v1.7.7
- **ORM:** GORM v1.23.3
- **Config:** Viper (config.yaml)
- **Logging:** Uber Zap (logs to /log/ directory)
- **Database:** MySQL 5.7+

## Database Tables

- `admin_user` - Admin users
- `admin_user_token` - Admin tokens
- `prometheus_alert` - Alerts