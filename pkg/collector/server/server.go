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
	DateFormat  string
	TLSCert     string
	TLSKey      string
	SigningKey  []byte
}

type Server struct {
	config  *Config
	storage StorageManager
	auth    TokenValidator
	server  *http.Server
}

func NewServer(cfg *Config, k8sClient kubernetes.Interface) (*Server, error) {
	storageManager, err := storage.NewManager(cfg.StoragePath, cfg.DateFormat)
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

func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
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

	// Extract metadata from headers
	namespace := r.Header.Get("X-PowerTool-Namespace")
	matchingLabels := r.Header.Get("X-PowerTool-Matching-Labels")
	powerToolName := r.Header.Get("X-PowerTool-Job-ID")
	filename := r.Header.Get("X-PowerTool-Filename")

	if namespace == "" || powerToolName == "" {
		http.Error(w, "Missing required headers", http.StatusBadRequest)
		return
	}

	if matchingLabels == "" {
		matchingLabels = "unknown"
	}

	if filename == "" {
		filename = fmt.Sprintf("%s.profile", powerToolName)
	}

	metadata := storage.ProfileMetadata{
		Namespace:     namespace,
		AppLabel:      matchingLabels,
		PowerToolName: powerToolName,
		Filename:      filename,
	}

	log.Printf("Authenticated request from %s for job %s, saving to %s/%s/%s",
		userInfo.Username, powerToolName, namespace, matchingLabels, powerToolName)

	if err := s.storage.SaveProfile(r.Body, metadata); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save profile: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
