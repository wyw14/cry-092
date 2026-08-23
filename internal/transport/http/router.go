package httptransport

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	assignmentapp "github.com/wyw14/cry-092/internal/application/assignment"
	handlingapp "github.com/wyw14/cry-092/internal/application/handling"
	queryapp "github.com/wyw14/cry-092/internal/application/query"
	responseapp "github.com/wyw14/cry-092/internal/application/response"
	reviewapp "github.com/wyw14/cry-092/internal/application/review"
	submissionapp "github.com/wyw14/cry-092/internal/application/submission"
	"github.com/wyw14/cry-092/internal/domain/response"
	"github.com/wyw14/cry-092/internal/domain/review"
	"github.com/wyw14/cry-092/internal/middleware"
	"go.uber.org/zap"
)

type Dependencies struct {
	Submission submissionapp.Service
	Assignment assignmentapp.Service
	Handling   handlingapp.Service
	Response   responseapp.Service
	Review     reviewapp.Service
	Query      queryapp.Service
	SigningKey []byte
	Issuer     string
	NewID      func() string
	Ready      func(context.Context) error
	Logger     *zap.Logger
}
type Handler struct {
	deps     Dependencies
	validate *validator.Validate
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	limiter := middleware.NewLimiter(20, 40)
	r.Use(middleware.RequestID(deps.NewID), middleware.Recover(deps.Logger), middleware.SecurityHeaders(), middleware.CORS(map[string]bool{"http://localhost:5173": true}), limiter.Middleware())
	h := &Handler{deps: deps, validate: validator.New()}
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", h.ready)
	api := r.Group("/api/v1", middleware.Authenticate(deps.SigningKey, deps.Issuer))
	api.POST("/proposals/:id/submit", middleware.RequireRoles("representative"), h.submit)
	api.POST("/proposals/:id/assign", middleware.RequireRoles("supervisor", "administrator"), h.assign)
	api.POST("/proposals/:id/accept", middleware.RequireRoles("unit_officer"), h.accept)
	api.POST("/proposals/:id/plans", middleware.RequireRoles("unit_officer"), h.startPlan)
	api.POST("/proposals/:id/replies", middleware.RequireRoles("unit_officer"), h.reply)
	api.POST("/proposals/:id/evaluations", middleware.RequireRoles("representative"), h.evaluate)
	api.GET("/proposals", h.listPlaceholder)
	return r
}

func (h *Handler) principal(c *gin.Context) middleware.Principal {
	p, _ := middleware.Current(c)
	return p
}
func (h *Handler) timed(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 8*time.Second)
}
func (h *Handler) ready(c *gin.Context) {
	ctx, cancel := h.timed(c)
	defer cancel()
	if err := h.deps.Ready(ctx); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) submit(c *gin.Context) {
	ctx, cancel := h.timed(c)
	defer cancel()
	snap, err := h.deps.Submission.Submit(ctx, c.Param("id"), h.principal(c).UserID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, snap)
}
func (h *Handler) assign(c *gin.Context) {
	ctx, cancel := h.timed(c)
	defer cancel()
	a, err := h.deps.Assignment.AutoAssign(ctx, c.Param("id"), h.principal(c).UserID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}
func (h *Handler) accept(c *gin.Context) {
	var in struct {
		UnitID string `json:"unit_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || h.validate.Struct(in) != nil {
		writeError(c, &validationError{"unit_id", "承办单位不能为空"})
		return
	}
	ctx, cancel := h.timed(c)
	defer cancel()
	if err := h.deps.Assignment.Accept(ctx, c.Param("id"), in.UnitID, h.principal(c).UserID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) startPlan(c *gin.Context) {
	var in struct {
		UnitID string   `json:"unit_id" validate:"required"`
		Steps  []string `json:"steps" validate:"min=1,dive,required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || h.validate.Struct(in) != nil {
		writeError(c, &validationError{"steps", "办理计划至少需要一个步骤"})
		return
	}
	ctx, cancel := h.timed(c)
	defer cancel()
	plan, err := h.deps.Handling.Start(ctx, c.Param("id"), in.UnitID, h.principal(c).UserID, in.Steps)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, plan)
}
func (h *Handler) reply(c *gin.Context) {
	var in struct {
		UnitID  string        `json:"unit_id" validate:"required"`
		Round   int           `json:"round" validate:"gte=1"`
		Kind    response.Kind `json:"kind" validate:"required"`
		Summary string        `json:"summary" validate:"required"`
		FileIDs []string      `json:"file_ids" validate:"min=1,dive,required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || h.validate.Struct(in) != nil {
		writeError(c, &validationError{"reply", "答复类型、说明和文件不能为空"})
		return
	}
	ctx, cancel := h.timed(c)
	defer cancel()
	reply, err := h.deps.Response.Submit(ctx, c.Param("id"), in.UnitID, h.principal(c).UserID, in.Round, in.Kind, in.Summary, in.FileIDs)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reply)
}
func (h *Handler) evaluate(c *gin.Context) {
	var in struct {
		ReplyID string        `json:"reply_id" validate:"required"`
		Rating  review.Rating `json:"rating" validate:"required"`
		Comment string        `json:"comment"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || h.validate.Struct(in) != nil {
		writeError(c, &validationError{"evaluation", "答复和评价不能为空"})
		return
	}
	ctx, cancel := h.timed(c)
	defer cancel()
	evaluation, err := h.deps.Review.Evaluate(ctx, c.Param("id"), in.ReplyID, h.principal(c).UserID, in.Rating, in.Comment)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, evaluation)
}
func (h *Handler) listPlaceholder(c *gin.Context) {
	page, pageErr := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, sizeErr := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageErr != nil || sizeErr != nil {
		writeError(c, &validationError{"pagination", "分页参数必须是整数"})
		return
	}
	principal := h.principal(c)
	filter := queryapp.ProposalFilter{
		Page: page, PageSize: size, Cursor: c.Query("cursor"),
		Sort: c.DefaultQuery("sort", "submitted_at"), Direction: c.DefaultQuery("direction", "desc"),
		Status: c.Query("status"), Category: c.Query("category"), UnitID: c.Query("unit_id"),
	}
	if principal.Roles["representative"] && !principal.Roles["supervisor"] && !principal.Roles["administrator"] {
		filter.OwnerID = principal.UserID
	}
	ctx, cancel := h.timed(c)
	defer cancel()
	result, err := h.deps.Query.List(ctx, filter)
	if err != nil {
		writeError(c, &validationError{"pagination", err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

type validationError struct{ field, message string }

func (e *validationError) Error() string { return e.message }
