package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kerezsiz42/scanner-operator2/frontend"
	"github.com/kerezsiz42/scanner-operator2/internal/oapi"
	"gorm.io/gorm"
)

type Server struct {
	db       *gorm.DB
	upgrader *websocket.Upgrader
}

func NewServer(db *gorm.DB) *Server {
	return &Server{
		db:       db,
		upgrader: &websocket.Upgrader{},
	}
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frontend.IndexHtml)
}

func (s *Server) GetBundleJs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frontend.BundleJs)
}

func (s *Server) GetOutputCss(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(frontend.OutputCss)
}

func (s *Server) GetHello(w http.ResponseWriter, r *http.Request) {
	resp := oapi.Hello{
		Hello: "world",
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetSubscribe(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		data := []byte("\"hello\"")
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Print("write:", err)
			break
		}

		time.Sleep(3 * time.Second)
	}
}
