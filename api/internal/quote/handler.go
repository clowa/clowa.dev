package quote

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler serves the HTTP endpoints for quotes. It depends only on the
// Repository interface, so swapping StaticRepository for a database-backed
// store requires no change here.
type Handler struct {
	repo Repository
}

// NewHandler wires a Handler to a Repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes mounts the quote endpoints onto the given router. The routes
// carry the public `/api` prefix because Caddy reverse-proxies `/api/*` to this
// service without stripping it (see docker/caddy/sites/clowa.dev.caddy).
func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/api/quote", h.getQuote)
}

// getQuote handles GET /api/quote and responds with a single quote as JSON.
func (h *Handler) getQuote(c *gin.Context) {
	q, err := h.repo.Random(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load quote"})
		return
	}

	c.JSON(http.StatusOK, q)
}
