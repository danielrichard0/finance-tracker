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

func New(cfg config.Config, expenseHandler *handler.ExpenseHandler, sqlDB *sql.DB) *gin.Engine {
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
		ex := v1.Group("/expenses")
		ex.GET("/", expenseHandler.ListExpenses)
		ex.POST("/", expenseHandler.CreateExpense)
		ex.GET("/:id", expenseHandler.GetExpenseByID)
		ex.PUT("/:id", expenseHandler.UpdateExpense)
		ex.DELETE("/:id", expenseHandler.DeleteExpense)
	}

	return r
}
