package httpapi

import (
	"net/http"

	"video-ops-agent/internal/agent/skills"

	"github.com/gin-gonic/gin"
)

type SkillHandler struct {
	service *skills.Service
}

func NewSkillHandler(service *skills.Service) *SkillHandler {
	return &SkillHandler{service: service}
}

func (h *SkillHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/skills")
	group.GET("", h.ListSkills)
	group.GET("/:id", h.GetSkill)
	group.POST("", h.CreateSkill)
	group.PUT("/:id", h.UpdateSkill)
	group.POST("/:id/enable", h.EnableSkill)
	group.POST("/:id/disable", h.DisableSkill)
}

func (h *SkillHandler) ListSkills(ctx *gin.Context) {
	skillList, err := h.service.List(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"skills": skillList})
}

func (h *SkillHandler) GetSkill(ctx *gin.Context) {
	skill, err := h.service.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"skill": skill})
}

func (h *SkillHandler) CreateSkill(ctx *gin.Context) {
	var req skills.DiagnosisSkill
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.service.Create(ctx.Request.Context(), req); err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	skill, err := h.service.Get(ctx.Request.Context(), req.ID)
	if err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"skill": skill})
}

func (h *SkillHandler) UpdateSkill(ctx *gin.Context) {
	var req skills.DiagnosisSkill
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.service.Update(ctx.Request.Context(), ctx.Param("id"), req); err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	skill, err := h.service.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"skill": skill})
}

func (h *SkillHandler) EnableSkill(ctx *gin.Context) {
	if err := h.service.Enable(ctx.Request.Context(), ctx.Param("id")); err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	skill, err := h.service.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"skill": skill})
}

func (h *SkillHandler) DisableSkill(ctx *gin.Context) {
	if err := h.service.Disable(ctx.Request.Context(), ctx.Param("id")); err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	skill, err := h.service.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		writeError(ctx, statusForSkillError(err), err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"skill": skill})
}

func statusForSkillError(err error) int {
	status := statusForStoreError(err)
	if status == http.StatusNotFound {
		return status
	}
	return http.StatusBadRequest
}
