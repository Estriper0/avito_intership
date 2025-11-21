package handlers

import (
	"errors"
	"net/http"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService service.IUserService
	validate    *validator.Validate
}

func NewUserHandler(g *gin.RouterGroup, userService service.IUserService, validate *validator.Validate) {
	r := &UserHandler{
		userService: userService,
		validate:    validate,
	}

	g.POST("/setIsActive", r.SetIsActive)
	g.GET("/getReview", r.GetReview)
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req dto.SetIsActiveRequest

	if err := c.Bind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, err)
		return
	}

	user, err := h.userService.SetIsActive(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(c, http.StatusNotFound, ErrStatusNotFound, err)
			return
		}
		respondWithError(c, http.StatusInternalServerError, ErrStatusInternal, err)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{
			"user": user,
		},
	)
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userId, ok := c.GetQuery("user_id")
	if !ok {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid query params"))
		return
	}

	pr, err := h.userService.GetReview(c.Request.Context(), userId)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, ErrStatusInternal, err)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{
			"user_id":       userId,
			"pull_requests": pr,
		},
	)
}
