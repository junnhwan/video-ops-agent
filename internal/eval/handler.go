package eval

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/eval")
	group.GET("/summary", h.GetSummary)
	group.GET("/skills/:id/summary", h.GetSkillSummary)
	group.POST("/runs", h.CreateRun)
	group.GET("/runs/:id", h.GetRun)
}

func (h *Handler) GetSummary(ctx *gin.Context) {
	summary, err := h.service.Summary(ctx.Request.Context(), SummaryFilter{})
	if err != nil {
		writeEvalError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"summary": summary})
}

func (h *Handler) GetSkillSummary(ctx *gin.Context) {
	summary, err := h.service.Summary(ctx.Request.Context(), SummaryFilter{SkillID: ctx.Param("id")})
	if err != nil {
		writeEvalError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"summary": summary})
}

func (h *Handler) CreateRun(ctx *gin.Context) {
	var req CreateRunInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeEvalError(ctx, http.StatusBadRequest, err)
		return
	}
	run, err := h.service.CreateRun(ctx.Request.Context(), req)
	if err != nil {
		writeEvalError(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"run": run})
}

func (h *Handler) GetRun(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		writeEvalError(ctx, http.StatusBadRequest, fmt.Errorf("id must be a positive integer"))
		return
	}
	run, err := h.service.GetRun(ctx.Request.Context(), uint(id))
	if err != nil {
		writeEvalError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"run": run})
}

func writeEvalError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, gin.H{"error": err.Error()})
}
