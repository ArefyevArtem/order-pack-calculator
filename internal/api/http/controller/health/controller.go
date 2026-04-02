package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Controller serves liveness and readiness probes.
type Controller struct {
	pool *pgxpool.Pool
}

// New builds a health controller; pool is used for readiness (DB ping).
func New(pool *pgxpool.Pool) *Controller {
	return &Controller{pool: pool}
}

// RegisterRoutes mounts liveness and readiness probes.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	r.GET("/live", c.liveness)
	r.GET("/ready", c.readiness)
}

func (c *Controller) liveness(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) readiness(ctx *gin.Context) {
	pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
	defer cancel()

	if err := c.pool.Ping(pingCtx); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": gin.H{"database": "unavailable"},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"checks": gin.H{"database": "ok"},
	})
}
