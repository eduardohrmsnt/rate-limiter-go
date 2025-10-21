package web

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type TestResponse struct {
	Message string `json:"message"`
	IP      string `json:"ip"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	ip := extractIP(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TestResponse{
		Message: "Request successful",
		IP:      ip,
	})
}
