package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

type ApiError struct {
	Error string
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/account", WrapToHandle(s.handle))

	log.Println("Running Server on port: ", s.listenAddr)

	log.Fatal(http.ListenAndServe(s.listenAddr, mux))
}

func (s *APIServer) handle(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	case "DELETE":
		return s.handleDeleteAccount(w, r)
	case "PUT":
		return s.handleUpdateAccount(w, r)
	}
	return fmt.Errorf("Unsupported method: %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	acc := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(acc); err != nil {
		return err
	}
	account := NewAccount(acc.Firstname, acc.Lastname)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, struct {
		ID         int
		Created_at time.Time
	}{
		ID:         account.ID,
		Created_at: account.CreatedAt,
	})
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func WrapToHandle(f func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
