package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	mu sync.Mutex
	store map[string]map[string]interface{}
}

func NewServer() *Server {
	s := &Server{store: make(map[string]map[string]interface{})}
	return s
}

func (s *Server) Start(port int) {
	r := http.NewServeMux()
	r.HandleFunc("/create", s.create)
	r.HandleFunc("/refund", s.refund)
	addr := ":"+strconv.Itoa(port)
	log.Printf("payments mock listening %s", addr)
	http.ListenAndServe(addr, r)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var req struct { Amount float64 `json:"amount"`; Currency string `json:"currency"`; VendorID string `json:"vendor_id"`; Memo string `json:"memo"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
	id := randID()
	s.mu.Lock()
	s.store[id] = map[string]interface{}{"payment_id":id, "amount":req.Amount, "currency":req.Currency, "status":"created"}
	s.mu.Unlock()
	json.NewEncoder(w).Encode(s.store[id])
}

func (s *Server) refund(w http.ResponseWriter, r *http.Request) {
	var req struct { PaymentID string `json:"payment_id"`; Reason string `json:"reason"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", 400); return }
	id := randID()
	s.mu.Lock()
	orig := s.store[req.PaymentID]
	s.mu.Unlock()
	if orig == nil { http.Error(w, "not found", 404); return }
	out := map[string]interface{}{"refund_id":id, "payment_id":req.PaymentID, "status":"refunded"}
	json.NewEncoder(w).Encode(out)
}

func randID() string { return fmt.Sprintf("p_%d", time.Now().UnixNano()+int64(rand.Intn(1000))) }
