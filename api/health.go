package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/pat"
	"github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/storage"
)

// HealthAPI provides health check and metrics endpoints
type HealthAPI struct {
	config    *config.Config
	startTime time.Time
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version,omitempty"`
}

// ReadinessResponse represents the readiness check response
type ReadinessResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Storage   string `json:"storage"`
	StorageOk bool   `json:"storage_ok"`
}

// CreateHealthAPI creates health check endpoints
func CreateHealthAPI(conf *config.Config, r *pat.Router, version string) *HealthAPI {
	healthAPI := &HealthAPI{
		config:    conf,
		startTime: time.Now(),
	}

	// Health check endpoint - basic liveness probe
	r.Path(conf.WebPath + "/health").Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		healthAPI.healthCheck(w, req, version)
	})

	// Readiness check endpoint - checks dependencies
	r.Path(conf.WebPath + "/ready").Methods("GET").HandlerFunc(healthAPI.readinessCheck)

	// Basic metrics endpoint in Prometheus format
	r.Path(conf.WebPath + "/metrics").Methods("GET").HandlerFunc(healthAPI.metrics)

	return healthAPI
}

func (h *HealthAPI) healthCheck(w http.ResponseWriter, req *http.Request, version string) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).String(),
		Version:   version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthAPI) readinessCheck(w http.ResponseWriter, req *http.Request) {
	storageOk := true
	storageType := "unknown"

	// Check storage connectivity
	if h.config.Storage != nil {
		storageType = h.config.StorageType
		// Try to get count to verify storage is working
		if _, err := h.config.Storage.List(0, 1); err != nil {
			storageOk = false
		}
	}

	status := "ready"
	statusCode := http.StatusOK
	if !storageOk {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	response := ReadinessResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Storage:   storageType,
		StorageOk: storageOk,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthAPI) metrics(w http.ResponseWriter, req *http.Request) {
	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get message count from storage
	var totalMessages int64
	if h.config.Storage != nil {
		totalMessages = int64(h.config.Storage.Count())
	}

	uptime := time.Since(h.startTime).Seconds()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	// Write Prometheus format metrics
	fmt.Fprintf(w, "# HELP mailhog_messages_total Total number of messages stored\n")
	fmt.Fprintf(w, "# TYPE mailhog_messages_total counter\n")
	fmt.Fprintf(w, "mailhog_messages_total %d\n", totalMessages)

	fmt.Fprintf(w, "# HELP mailhog_uptime_seconds Time since MailHog started in seconds\n")
	fmt.Fprintf(w, "# TYPE mailhog_uptime_seconds gauge\n")
	fmt.Fprintf(w, "mailhog_uptime_seconds %.2f\n", uptime)

	fmt.Fprintf(w, "# HELP mailhog_memory_usage_bytes Current memory usage in bytes\n")
	fmt.Fprintf(w, "# TYPE mailhog_memory_usage_bytes gauge\n")
	fmt.Fprintf(w, "mailhog_memory_usage_bytes %d\n", m.Alloc)

	fmt.Fprintf(w, "# HELP mailhog_goroutines Number of goroutines currently running\n")
	fmt.Fprintf(w, "# TYPE mailhog_goroutines gauge\n")
	fmt.Fprintf(w, "mailhog_goroutines %d\n", runtime.NumGoroutine())

	// Add storage type as a label
	fmt.Fprintf(w, "# HELP mailhog_storage_type_info Storage type information\n")
	fmt.Fprintf(w, "# TYPE mailhog_storage_type_info gauge\n")
	fmt.Fprintf(w, "mailhog_storage_type_info{storage_type=\"%s\"} 1\n", h.config.StorageType)

	// Add storage connectivity status
	fmt.Fprintf(w, "# HELP mailhog_storage_connected Storage connection status\n")
	fmt.Fprintf(w, "# TYPE mailhog_storage_connected gauge\n")
	if h.config.Storage != nil {
		if mongoStorage, ok := h.config.Storage.(*storage.MongoDB); ok {
			_ = mongoStorage // Could add MongoDB-specific checks here
		}
		fmt.Fprintf(w, "mailhog_storage_connected 1\n")
	} else {
		fmt.Fprintf(w, "mailhog_storage_connected 0\n")
	}
}
