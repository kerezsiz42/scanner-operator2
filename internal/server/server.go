package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	root "github.com/kerezsiz42/scanner-operator2"
	"github.com/kerezsiz42/scanner-operator2/internal/oapi"
)

var upgrader = websocket.Upgrader{}

type Server struct{}

func NewServer() (*Server, *http.ServeMux) {
	// Serving static files of the UI
	mux := http.NewServeMux()
	mux.Handle("/bundle.js", http.FileServer(http.FS(root.BundleJs)))
	mux.Handle("/", http.FileServer(http.FS(root.IndexHtml)))
	mux.Handle("/output.css", http.FileServer(http.FS(root.OutputCss)))
	mux.HandleFunc("/subscribe", subscribe)

	return &Server{}, mux
}

func (Server) GetHello(w http.ResponseWriter, r *http.Request) {
	resp := oapi.Hello{
		Hello: "world",
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		data := []byte("hello")
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Print("write:", err)
			break
		}

		time.Sleep(3 * time.Second)
	}
}
