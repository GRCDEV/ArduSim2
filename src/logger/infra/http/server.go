package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/GRCDEV/ArduSim2/logger/usecases"
)

type Server struct {
	server       *http.Server
	downloadUC   *usecases.DownloadLogsUseCase
}

func NewServer(address string, downloadUC *usecases.DownloadLogsUseCase) *Server {
	s := &Server{
		downloadUC: downloadUC,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/logs/download", s.handleDownload)

	s.server = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	fmt.Printf("[HTTP Server] Listening on %s\n", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"logs.zip\"")

	err := s.downloadUC.Execute(r.Context(), w)
	if err != nil {
		fmt.Printf("[HTTP Server] Error generating zip: %v\n", err)
		// We can't write http.Error after writing headers, but we can log it.
	}
}
