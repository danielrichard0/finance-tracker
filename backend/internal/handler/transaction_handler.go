package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"expense-tracker/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	service service.TransactionService
}

type saveTransactionRequest struct {
	Title           string  `json:"title" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	Category        string  `json:"category"`
	Notes           string  `json:"notes"`
	TransactionDate string  `json:"transaction_date"`
	Type            string  `json:"type" binding:"required"`
	UserID          string  `json:"user_id" binding:"required"`
}

func NewTransactionHandler(service service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req saveTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := toSaveInput(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, "transaction_date must be YYYY-MM-DD")
		return
	}

	transaction, err := h.service.CreateTransaction(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "title must not be empty and amount must be greater than zero")
			return
		}
		if errors.Is(err, service.ErrInvalidTransactionType) {
			respondError(c, http.StatusBadRequest, "type must be E or I")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create transaction")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": transaction})
}

func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	limit := readQueryInt(c, "limit", 20)
	offset := readQueryInt(c, "offset", 0)

	transactions, err := h.service.ListTransactions(c.Request.Context(), service.ListTransactionsInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	id, ok := readParamID(c)
	if !ok {
		return
	}

	transaction, err := h.service.GetTransactionByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "transaction not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to fetch transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transaction})
}

func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	id, ok := readParamID(c)
	if !ok {
		return
	}

	var req saveTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := toSaveInput(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, "transaction_date must be YYYY-MM-DD")
		return
	}

	transaction, err := h.service.UpdateTransaction(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "transaction not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "title must not be empty and amount must be greater than zero")
			return
		}
		if errors.Is(err, service.ErrInvalidTransactionType) {
			respondError(c, http.StatusBadRequest, "type must be E or I")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to update transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transaction})
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	id, ok := readParamID(c)
	if !ok {
		return
	}

	err := h.service.DeleteTransaction(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "transaction not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete transaction")
		return
	}

	c.Status(http.StatusNoContent)
}

func toSaveInput(req saveTransactionRequest) (service.SaveTransactionInput, error) {
	var transactionDate time.Time
	var err error

	if req.TransactionDate != "" {
		transactionDate, err = time.Parse("2006-01-02", req.TransactionDate)
		if err != nil {
			return service.SaveTransactionInput{}, err
		}
	}

	return service.SaveTransactionInput{
		Title:           req.Title,
		Amount:          req.Amount,
		Category:        req.Category,
		Notes:           req.Notes,
		TransactionDate: transactionDate,
		Type:            req.Type,
	}, nil
}

func readParamID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func readQueryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
