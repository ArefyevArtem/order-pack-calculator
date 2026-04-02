package calculator

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"order-pack-calculator/internal/domain"
	svcalc "order-pack-calculator/internal/usecase/calculator"
)

// UseCase is the behavior this controller depends on — interface defined at the consumer (handler package), per Go practice.
type UseCase interface {
	Calculate(ctx context.Context, items int) (*svcalc.Result, error)
	ReplacePackSizes(ctx context.Context, sizes []int) error
	ListPackSizes(ctx context.Context) ([]int, error)
}

// Controller exposes HTTP routes for pack configuration and calculation.
type Controller struct {
	uc  UseCase
	log *slog.Logger
}

// New creates a calculator controller.
func New(uc UseCase, log *slog.Logger) *Controller {
	return &Controller{uc: uc, log: log}
}

// RegisterRoutes mounts /api/v1 routes (PUT replaces the full pack size list).
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	
	api.PUT("/pack-sizes", c.replacePackSizes)
	api.POST("/calculate", c.calculate)
}

func (c *Controller) replacePackSizes(ctx *gin.Context) {
	var req PackSizesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("pack-sizes bind failed", "error", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := c.uc.ReplacePackSizes(ctx.Request.Context(), req.Sizes); err != nil {
		switch {
		case errors.Is(err, svcalc.ErrNoPackSizes):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.log.Warn("pack-sizes rejected", "error", err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		}
		return
	}
	stored, err := c.uc.ListPackSizes(ctx.Request.Context())
	if err != nil {
		c.log.Error("list pack sizes after save", "error", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, PackSizesResponse{Sizes: stored})
}

func (c *Controller) calculate(ctx *gin.Context) {
	var req CalculateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.log.Warn("calculate bind failed", "error", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request: " + err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		c.log.Warn("calculate validation failed", "error", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	res, err := c.uc.Calculate(ctx.Request.Context(), req.Items)
	if err != nil {
		switch {
		case errors.Is(err, svcalc.ErrNoPackSizes):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrNoExactPacking):
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.log.Error("calculate failed", "error", err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, CalculateResponse{
		Packs:   toPackLines(res.Packs),
		Message: "ok",
	})
}

func toPackLines(in []svcalc.PackLine) []PackLine {
	out := make([]PackLine, 0, len(in))
	for _, l := range in {
		out = append(out, PackLine{Pack: l.Pack, Quantity: l.Quantity})
	}
	return out
}
