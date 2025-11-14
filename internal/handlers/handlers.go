package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/iben12/counter-app/internal/models"
)

type Server struct {
    db *pgxpool.Pool
}

func NewRouter(pool *pgxpool.Pool) http.Handler {
	s := &Server{db: pool}
	r := mux.NewRouter()
	r.HandleFunc("/health", s.health).Methods("GET")

	r.HandleFunc("/counters", s.listCounters).Methods("GET")
	r.HandleFunc("/counters", s.createCounter).Methods("POST")
	r.HandleFunc("/counters/{id}", s.getCounter).Methods("GET")
	r.HandleFunc("/counters/{id}/frequency", s.updateCounterFrequency).Methods("POST")

	// Count endpoints
	r.HandleFunc("/counters/{id}/count", s.getCurrentCount).Methods("GET")
	r.HandleFunc("/counters/{id}/count/increment", s.incrementCount).Methods("POST")
	r.HandleFunc("/counters/{id}/count/decrement", s.decrementCount).Methods("POST")
	r.HandleFunc("/counters/{id}/counts", s.getCountHistory).Methods("GET")

	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) listCounters(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    cs, err := models.GetAllCounters(ctx, s.db)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(cs)
}

type createReq struct {
    Name      string `json:"name"`
    Frequency string `json:"frequency,omitempty"`
}

func (s *Server) createCounter(w http.ResponseWriter, r *http.Request) {
    var req createReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    if req.Name == "" {
        http.Error(w, "name required", http.StatusBadRequest)
        return
    }
    c, err := models.CreateCounter(r.Context(), s.db, req.Name, req.Frequency)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(c)
}

func (s *Server) getCounter(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    c, err := models.GetCounterByID(r.Context(), s.db, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(c)
}

type incFreqReq struct {
    Frequency string `json:"frequency"`
}

func (s *Server) updateCounterFrequency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req incFreqReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Frequency == "" {
		http.Error(w, "frequency required", http.StatusBadRequest)
		return
	}
	c, err := models.UpdateCounterFrequency(r.Context(), s.db, id, req.Frequency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

type incCountReq struct {
	Delta int64 `json:"delta"`
}

func (s *Server) getCurrentCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	cnt, err := models.GetOrCreateCurrentCount(r.Context(), s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cnt)
}

func (s *Server) incrementCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req incCountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// default delta to 1 if no body
		req.Delta = 1
	}
	if req.Delta == 0 {
		req.Delta = 1
	}
	cnt, err := models.IncrementCurrentCount(r.Context(), s.db, id, req.Delta)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cnt)
}

func (s *Server) decrementCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req incCountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// default delta to 1 if no body
		req.Delta = 1
	}
	if req.Delta == 0 {
		req.Delta = 1
	}
	// Use negative delta to decrement
	cnt, err := models.IncrementCurrentCount(r.Context(), s.db, id, -req.Delta)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cnt)
}

func (s *Server) getCountHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	counts, err := models.GetCountHistory(r.Context(), s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(counts)
}
