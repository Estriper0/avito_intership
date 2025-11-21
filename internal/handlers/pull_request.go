package handlers

import (
	"errors"
	"net/http"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PullRequestHandler struct {
	prService service.IPullRequestService
	validate  *validator.Validate
}

func NewPullRequestHandler(g *gin.RouterGroup, prService service.IPullRequestService, validate *validator.Validate) {
	r := &PullRequestHandler{
		prService: prService,
		validate:  validate,
	}

	g.POST("/create", r.Create)
	g.POST("/merge", r.Merge)
	g.POST("/reassign", r.Reassign)
}

func (h *PullRequestHandler) Create(c *gin.Context) {
	var req dto.PrCreateRequest

	if err := c.Bind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, err)
		return
	}

	pr, err := h.prService.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrPullRequestALreadyExists) {
			respondWithError(c, http.StatusConflict, ErrStatusPrExists, err)
			return
		} else if errors.Is(err, service.ErrNotFound) {
			respondWithError(c, http.StatusNotFound, ErrStatusNotFound, err)
			return
		}
		respondWithError(c, http.StatusInternalServerError, ErrStatusInternal, err)
		return
	}
	c.JSON(
		http.StatusCreated,
		gin.H{
			"pr": pr,
		},
	)
}

func (h *PullRequestHandler) Merge(c *gin.Context) {
	var req dto.MergeRequest

	if err := c.Bind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, err)
		return
	}

	pr, err := h.prService.Merge(c.Request.Context(), req.PrId)
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
			"pr": pr,
		},
	)
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req dto.ReassignRequest

	if err := c.Bind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, err)
		return
	}

	resp, err := h.prService.Reassign(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(c, http.StatusNotFound, ErrStatusNotFound, err)
			return
		} else if errors.Is(err, service.ErrPullRequestMerged) {
			respondWithError(c, http.StatusConflict, ErrStatusPrMerged, err)
			return
		} else if errors.Is(err, service.ErrNoCandidate) {
			respondWithError(c, http.StatusNotFound, ErrStatusNoCandidate, err)
			return
		}
		respondWithError(c, http.StatusInternalServerError, ErrStatusInternal, err)
		return
	}

	c.JSON(
		http.StatusOK,
		resp,
	)
}
