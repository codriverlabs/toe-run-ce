package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"toe/pkg/collector/auth"
	"toe/pkg/collector/storage"

	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Port        int
	StoragePath string
	TLSCert     string
	TLSKey      string
	SigningKey  []byte
}

type Server struct {
	config  *Config
	storage *storage.Manager
	auth    *auth.K8sTokenValidator
	server  *http.Server
}

func NewServer(cfg *Config, k8sClient kubernetes.Interface) (*Server, error) {
	storageManager, err := storage.NewManager(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage manager: %w", err)
	}

	s := &Server{
		config:  cfg,
		storage: storageManager,
		auth:    auth.NewK8sTokenValidator(k8sClient, "toe-sdk-collector"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/profile", s.handleProfile)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return s, nil
}

func (s *Server) Start() error {
	log.Printf("Starting server on port %d", s.config.Port)
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		return s.server.ListenAndServeTLS(s.config.TLSCert, s.config.TLSKey)
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() {
	if err := s.server.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	userInfo, err := s.auth.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
		return
	}

	// Extract job ID from request headers or URL
	jobID := r.Header.Get("X-PowerTool-Job-ID")
	if jobID == "" {
		http.Error(w, "Missing X-PowerTool-Job-ID header", http.StatusBadRequest)
		return
	}

	// Log successful authentication
	log.Printf("Authenticated request from %s for job %s", userInfo.Username, jobID)

	if err := s.storage.SaveProfile(r.Body, jobID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save profile: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
