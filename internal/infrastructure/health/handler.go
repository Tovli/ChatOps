package health

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.Logger
	db     *sql.DB
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func NewHandler(logger *zap.Logger, db *sql.DB) *Handler {
	return &Handler{
		logger: logger,
		db:     db,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]string)

	// Check database
	if err := h.db.PingContext(r.Context()); err != nil {
		services["database"] = "unhealthy"
	} else {
		services["database"] = "healthy"
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  services,
	}

	// If any service is unhealthy, mark overall status as unhealthy
	for _, status := range services {
		if status != "healthy" {
			response.Status = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("alive"))
}
