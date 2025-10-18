package files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type Server struct {
	mu sync.Mutex
	store map[string]string
	base string
}

func NewServer() *Server {
	b := os.TempDir()
	s := &Server{store: make(map[string]string), base: b}
	return s
}

func (s *Server) Start(port int) {
	r := http.NewServeMux()
	r.HandleFunc("/read", s.read)
	r.HandleFunc("/write", s.write)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("files mock listening %s", addr)
	http.ListenAndServe(addr, r)
}

func (s *Server) read(w http.ResponseWriter, r *http.Request) {
	var req struct{ Path string `json:"path"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
	// in-memory or tmp
	s.mu.Lock()
	content, ok := s.store[req.Path]
	s.mu.Unlock()
	if ok {
		json.NewEncoder(w).Encode(map[string]string{"path":req.Path, "content":content})
		return
	}
	// try disk
	p := filepath.Join(s.base, filepath.FromSlash(req.Path))
	b, err := ioutil.ReadFile(p)
	if err != nil { http.Error(w, "not found", 404); return }
	json.NewEncoder(w).Encode(map[string]string{"path":req.Path, "content":string(b)})
}

func (s *Server) write(w http.ResponseWriter, r *http.Request) {
	var req struct{ Path string `json:"path"`; Content string `json:"content"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
	s.mu.Lock()
	s.store[req.Path] = req.Content
	s.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]string{"path":req.Path, "status":"written"})
}
