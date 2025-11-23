package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/Estriper0/avito_intership/internal/handlers/dto"
	"github.com/Estriper0/avito_intership/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	queueCap    int = 10
	countWorker int = 5
)

type Task struct {
	Id       uuid.UUID
	TeamName string
}

type PullRequestHandler struct {
	prService service.IPullRequestService
	taskQueue chan Task
	validate  *validator.Validate
}

func NewPullRequestHandler(g *gin.RouterGroup, prService service.IPullRequestService, validate *validator.Validate) {
	r := &PullRequestHandler{
		prService: prService,
		taskQueue: make(chan Task, queueCap),
		validate:  validate,
	}

	//Start workers to handle heavy tasks 
	startWorkers(r, countWorker)

	g.POST("/create", r.Create)
	g.POST("/merge", r.Merge)
	g.POST("/reassign", r.Reassign)
	g.POST("/reassign/team", r.ReassignAllInactiveReviewersByTeam)
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

func (h *PullRequestHandler) ReassignAllInactiveReviewersByTeam(c *gin.Context) {
	teamName, ok := c.GetQuery("team_name")
	if !ok {
		respondWithError(c, http.StatusBadRequest, ErrStatusBadRequest, errors.New("invalid query params"))
		return
	}

	go func() {
		h.taskQueue <- Task{Id: uuid.New(), TeamName: teamName}
	}()

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "The task has been received",
		},
	)
}

func startWorkers(h *PullRequestHandler, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func() {
			for task := range h.taskQueue {
				h.prService.ReassignAllInactiveReviewersByTeam(context.Background(), task.TeamName)
			}
		}()
	}
}
