# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **paas-finops-extend**, a Go + Gin implementation of a FinOps API backend with admin user management and alert functionality.

## Build and Run Commands

```bash
# Build
go build -o server main.go

# Run
./server

# Dependencies
go mod tidy -compat=1.17
```

The server runs on port 8888 (configurable in `config.yaml`). Requires MySQL 5.7+ with schema from `/static-files/finops_schema.sql`.

## Architecture

**Request Flow:**
```
Routes (router/) → API Handlers (api/v1/) → Services (service/) → Models (model/) → GORM → MySQL
```

**Key Directories:**
- `api/v1/manage/` - Admin panel API handlers
- `service/manage/` - Business logic layer
- `model/manage/` - GORM models with request structs
- `router/manage/` - Route registration
- `middleware/` - JWT auth and CORS middleware
- `initialize/` - Database and router initialization
- `global/` - Global variables (GVA_DB, GVA_LOG, GVA_CONFIG)

**API Group:**
- `/manage-api/v1/` - Admin routes, protected by `AdminJWTAuth()`

## Available APIs

**Admin User:**
- `POST /manage-api/v1/adminUser/login` - Admin login
- `POST /manage-api/v1/createFinopsAdminUser` - Create admin user
- `PUT /manage-api/v1/adminUser/name` - Update admin name
- `PUT /manage-api/v1/adminUser/password` - Update password
- `GET /manage-api/v1/adminUser/profile` - Get admin profile
- `DELETE /manage-api/v1/logout` - Logout
- `POST /manage-api/v1/upload/file` - Upload file

**Alerts:**
- `POST /manage-api/v1/alerts` - Create alert
- `GET /manage-api/v1/alerts` - Get alert list
- `GET /manage-api/v1/alerts/:alertId` - Get alert by ID
- `PUT /manage-api/v1/alerts/:alertId` - Update alert
- `DELETE /manage-api/v1/alerts/:alertId` - Delete alert
- `DELETE /manage-api/v1/alerts` - Batch delete alerts

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
```

**Model Definition:**
```go
type FinopsAdminUser struct {
    AdminUserId int `json:"adminUserId" gorm:"primarykey;AUTO_INCREMENT"`
    // ...
}
func (FinopsAdminUser) TableName() string {
    return "finops_admin_user"
}
```

**Database Access:**
```go
global.GVA_DB.Where("condition").First(&result)
```

## Tech Stack

- **Framework:** Gin v1.7.7
- **ORM:** GORM v1.23.3
- **Config:** Viper (config.yaml)
- **Logging:** Uber Zap (logs to /log/ directory)
- **Database:** MySQL 5.7+

## Database Tables

- `finops_admin_user` - Admin users
- `finops_admin_user_token` - Admin tokens
- `finops_alert` - Alerts

## Test Credentials

Admin panel: `admin` / `123456`
