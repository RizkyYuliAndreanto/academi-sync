# Backend (Go + Gin)

Minimal starter for the Go API using Gin.

Run instructions (after installing Go and dependencies):

1. Initialize module (if not done):
   go mod init bimbingan-backend
2. Add Gin dependency:
   go get github.com/gin-gonic/gin
3. Build/run:
   go run main.go

Health endpoint: GET /health

This folder contains a minimal server skeleton, graceful shutdown and basic middleware as a starting point for TASK-011.