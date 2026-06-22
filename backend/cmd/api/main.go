package main

import (
	"log"
	"net/http"
	"time"

	"expense-tracker/backend/internal/config"
	"expense-tracker/backend/internal/db"
	"expense-tracker/backend/internal/handler"
	"expense-tracker/backend/internal/repository"
	"expense-tracker/backend/internal/router"
	"expense-tracker/backend/internal/service"
)

func main() {
	cfg := config.Load()

	sqlDB, err := db.NewMySQL(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	transactionRepo := repository.NewTransactionRepository(sqlDB)
	transactionService := service.NewTransactionService(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	userRepo := repository.NewUserRepository(sqlDB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	engine := router.New(cfg, transactionHandler, userHandler, sqlDB)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("backend running on http://localhost:%s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
