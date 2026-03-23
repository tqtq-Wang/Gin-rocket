package handler

import (
	"errors"
	"net/http"
	"strconv"

	"gin-rocket/internal/service"
	"gin-rocket/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetByID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "user not found")
			return
		}

		response.Fail(c, http.StatusInternalServerError, "query user failed")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"user": user})
}
