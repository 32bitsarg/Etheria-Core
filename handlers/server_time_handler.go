package handlers

import (
	"net/http"
	"server-backend/config"
	"time"

	"github.com/gin-gonic/gin"
)

func ServerTimeHandler(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cargando configuración"})
		return
	}
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	c.JSON(http.StatusOK, gin.H{
		"server_time": now.Format(time.RFC3339),
	})
}
