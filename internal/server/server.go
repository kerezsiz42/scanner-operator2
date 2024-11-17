package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/kerezsiz42/scanner-operator2/frontend"
	"github.com/kerezsiz42/scanner-operator2/internal/oapi"
	"github.com/kerezsiz42/scanner-operator2/internal/service"
	"gorm.io/gorm"
)

type Server struct {
	scanService service.ScanServiceInterface
	upgrader    *websocket.Upgrader
	logger      logr.Logger
	broadcastCh chan string
	connections map[*websocket.Conn]chan string
	mu          sync.Mutex
	once        sync.Once
}

func NewServer(
	scanService service.ScanServiceInterface,
	logger logr.Logger,
) *Server {
	return &Server{
		upgrader:    &websocket.Upgrader{},
		scanService: scanService,
		logger:      logger,
		broadcastCh: make(chan string),
		connections: make(map[*websocket.Conn]chan string),
		mu:          sync.Mutex{},
	}
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	defer observeDuration("GET", "/")()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frontend.IndexHtml)
}

func (s *Server) GetBundleJs(w http.ResponseWriter, r *http.Request) {
	defer observeDuration("GET", "/bundle.js")()
	w.Header().Set("Content-Type", "text/javascript")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frontend.BundleJs)
}

func (s *Server) GetOutputCss(w http.ResponseWriter, r *http.Request) {
	defer observeDuration("GET", "/output.css")()
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frontend.OutputCss)
}

func (s *Server) GetSubscribe(w http.ResponseWriter, r *http.Request) {
	defer observeDuration("GET", "/subscribe")()
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error(err, "GetSubscribe")
		return
	}

	s.once.Do(func() {
		go func() {
			for {
				message := <-s.broadcastCh

				for _, ch := range s.connections {
					ch <- message
				}
			}
		}()
	})

	defer c.Close()

	ch := make(chan string)
	defer close(ch)

	s.mu.Lock()
	s.connections[c] = ch
	s.mu.Unlock()

	go func() {
		for {
			imageId, ok := <-ch
			if !ok {
				return
			}

			data := []byte("\"" + imageId + "\"")
			if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
				s.logger.Error(err, "Websocket")
			}
		}
	}()

	for {
		msgType, _, err := c.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.CloseMessage {
			break
		}
	}

	s.mu.Lock()
	delete(s.connections, c)
	s.mu.Unlock()
}

func (s *Server) GetScanResults(w http.ResponseWriter, r *http.Request) {
	defer observeDuration("GET", "/scan-results")()
	scanResults, err := s.scanService.ListScanResults()
	if err != nil {
		s.logger.Error(err, "GetScanResults")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := []oapi.ScanResult{}
	for _, scanResult := range scanResults {
		res = append(res, oapi.ScanResult{
			ImageId: scanResult.ImageID,
			Report:  json.RawMessage(scanResult.Report),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.Error(err, "GetScanResults")
	}
}

func (s *Server) PutScanResults(w http.ResponseWriter, r *http.Request) {
	defer observeDuration("PUT", "/scan-results")()
	oapiScanResult := oapi.ScanResult{}
	if err := json.NewDecoder(r.Body).Decode(&oapiScanResult); err != nil {
		s.logger.Error(err, "PutScanResults")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	scanResult, err := s.scanService.UpsertScanResult(oapiScanResult.ImageId, string(oapiScanResult.Report))
	if errors.Is(err, service.InvalidCycloneDXBOM) {
		s.logger.Error(err, "PutScanResults")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	} else if err != nil {
		s.logger.Error(err, "PutScanResults")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.broadcastCh <- scanResult.ImageID
	s.logger.Info("PutScanResults", "new imageId broadcasted", scanResult.ImageID)

	res := oapi.ScanResult{
		ImageId: scanResult.ImageID,
		Report:  json.RawMessage(scanResult.Report),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.Error(err, "PutScanResults")
	}
}

func (s *Server) DeleteScanResultsImageId(w http.ResponseWriter, r *http.Request, imageId string) {
	defer observeDuration("DELETE", "/scan-results/{imageId}")()
	if err := s.scanService.DeleteScanResult(imageId); err != nil {
		s.logger.Error(err, "DeleteScanResultsImageId")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetScanResultsImageId(w http.ResponseWriter, r *http.Request, imageId string) {
	defer observeDuration("GET", "/scan-results/{imageId}")()
	scanResult, err := s.scanService.GetScanResult(imageId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error(err, "GetScanResultsImageId")
		http.Error(w, "Not Found", http.StatusNotFound)
	} else if err != nil {
		s.logger.Error(err, "GetScanResultsImageId")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := oapi.ScanResult{
		ImageId: scanResult.ImageID,
		Report:  json.RawMessage(scanResult.Report),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.logger.Error(err, "GetScanResultsImageId")
	}
}
