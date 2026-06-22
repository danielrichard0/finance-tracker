package handler

import (
	"errors"
	"net/http"

	"expense-tracker/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

type saveUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Username string `json:"username" binding:"required"`
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req saveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), service.SaveUserInput{
		Email:    req.Email,
		Phone:    req.Phone,
		Username: req.Username,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "email and username are required")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create user : "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	limit := readQueryInt(c, "limit", 20)
	offset := readQueryInt(c, "offset", 0)

	users, err := h.service.ListUsers(c.Request.Context(), service.ListUsersInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "user not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req saveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, service.SaveUserInput{
		Email:    req.Email,
		Phone:    req.Phone,
		Username: req.Username,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "email and username are required")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	// if !ok {
	// 	return
	// }

	err := h.service.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(c, http.StatusNotFound, "user not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete user")
		return
	}

	c.Status(http.StatusNoContent)
}
