package handlers

import (
	"errors"
	"net/http"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TeamHandler struct {
	teamService service.ITeamService
	validate    *validator.Validate
}

func NewTeamHandler(g *gin.RouterGroup, teamService service.ITeamService, validate *validator.Validate) {
	r := &TeamHandler{
		teamService: teamService,
		validate:    validate,
	}

	g.POST("/add", r.Add)
	g.GET("/get", r.Get)
}

func (h *TeamHandler) Add(c *gin.Context) {
	var req dto.Team

	if err := c.Bind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid request body"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, err)
		return
	}

	_, err := h.teamService.Add(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrTeamAlreadyExists) {
			respondWithError(c, http.StatusBadRequest, ErrStatusTeamExists, err)
			return
		}
		respondWithError(c, http.StatusInternalServerError, ErrStatusInternal, err)
		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"team": req,
		},
	)
}

func (h *TeamHandler) Get(c *gin.Context) {
	teamName, ok := c.GetQuery("team_name")
	if !ok {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid query params"))
		return
	}

	team, err := h.teamService.Get(c.Request.Context(), teamName)
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
		team,
	)
}
