package router

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"expense-tracker/backend/internal/config"
	"expense-tracker/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, transactionHandler *handler.TransactionHandler, sqlDB *sql.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"error":  "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"port":   cfg.ServerPort,
		})
	})

	v1 := r.Group("/api/v1")
	{
		tx := v1.Group("/transactions")
		tx.GET("/", transactionHandler.ListTransactions)
		tx.POST("/", transactionHandler.CreateTransaction)
		tx.GET("/:id", transactionHandler.GetTransactionByID)
		tx.PUT("/:id", transactionHandler.UpdateTransaction)
		tx.DELETE("/:id", transactionHandler.DeleteTransaction)
	}

	return r
}
