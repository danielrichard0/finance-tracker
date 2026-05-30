package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"expense-tracker/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ExpenseHandler struct {
	service service.ExpenseService
}

type saveExpenseRequest struct {
	Title       string  `json:"title" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Category    string  `json:"category"`
	Notes       string  `json:"notes"`
	ExpenseDate string  `json:"expense_date"`
}

func NewExpenseHandler(service service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	var req saveExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := toSaveInput(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, "expense_date must be YYYY-MM-DD")
		return
	}

	expense, err := h.service.CreateExpense(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "title must not be empty and amount must be greater than zero")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create expense")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": expense})
}

func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	limit := readQueryInt(c, "limit", 20)
	offset := readQueryInt(c, "offset", 0)

	expenses, err := h.service.ListExpenses(c.Request.Context(), service.ListExpensesInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list expenses")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": expenses})
}

func (h *ExpenseHandler) GetExpenseByID(c *gin.Context) {
	id, ok := readParamID(c)
	if !ok {
		return
	}

	expense, err := h.service.GetExpenseByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "expense not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to fetch expense")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": expense})
}

func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	id, ok := readParamID(c)
	if !ok {
		return
	}

	var req saveExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := toSaveInput(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, "expense_date must be YYYY-MM-DD")
		return
	}

	expense, err := h.service.UpdateExpense(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "expense not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "title must not be empty and amount must be greater than zero")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to update expense")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": expense})
}

func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	id, ok := readParamID(c)
	if !ok {
		return
	}

	err := h.service.DeleteExpense(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "expense not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete expense")
		return
	}

	c.Status(http.StatusNoContent)
}

func toSaveInput(req saveExpenseRequest) (service.SaveExpenseInput, error) {
	var expenseDate time.Time
	var err error

	if req.ExpenseDate != "" {
		expenseDate, err = time.Parse("2006-01-02", req.ExpenseDate)
		if err != nil {
			return service.SaveExpenseInput{}, err
		}
	}

	return service.SaveExpenseInput{
		Title:       req.Title,
		Amount:      req.Amount,
		Category:    req.Category,
		Notes:       req.Notes,
		ExpenseDate: expenseDate,
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
