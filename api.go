package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gofrs/uuid"
)

type ServerConf struct {
	Addr string
	Port string
}

type Runnable interface {
	Run()
}

// NewServer web server constructor
func NewServer(mgr Manager, cfg ServerConf) (Runnable, error) {
	return &server{
		port:    cfg.Port,
		manager: mgr,
	}, nil
}

type updateRequest struct {
	State         GameState `json:"state"`
	Amount        float64   `json:"amount"`
	TransactionID uuid.UUID `json:"transactionId"`
}

type server struct {
	manager Manager
	port    string
}

// Run function register all server routes and starts the server
func (s *server) Run() {

	Log("Start HTTP server on port %v", s.port)
	// TODO: use gorilla/mux here in future if needed
	http.HandleFunc("/state", s.stateHandler)

	http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
}

func (s server) stateHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	updateSourceType := sourceType(req.Header.Get(sourceTypeHeaderName))
	_, ok := sourceTypes[updateSourceType]
	if !ok {
		http.Error(w, "Unknown source type", http.StatusBadRequest)
		return
	}

	// TODO: move to common handler registration func
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	update := updateRequest{}
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, fmt.Sprintf("Incorrect request body %v", err), http.StatusBadRequest)
		return
	}

	if err = s.manager.Update(req.Context(), User{ID: defaultUserUUID}, convertToBalanceUpdate(update)); err != nil {
		// TODO: add custom error type to diversify http error codes (now it's only 500)
		http.Error(w, fmt.Sprintf("Failed to proceed balance update request %v", err), http.StatusInternalServerError)
		return
	}

}

func convertToBalanceUpdate(update updateRequest) BalanceUpdate {
	return BalanceUpdate{
		ID:        update.TransactionID,
		Amount:    update.Amount,
		GameState: update.State,
	}
}
