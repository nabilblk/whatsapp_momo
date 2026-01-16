package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) Handle(c *gin.Context) {
	// Check database connectivity
	dbConnected := true
	if err := h.db.Ping(); err != nil {
		dbConnected = false
	}

	uptimeSeconds := int64(time.Since(h.startTime).Seconds())

	status := "healthy"
	httpStatus := http.StatusOK
	if !dbConnected {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":             status,
		"db_connected":       dbConnected,
		"whatsapp_connected": true, // This would be updated by bridge health check
		"uptime_seconds":     uptimeSeconds,
	})
}
