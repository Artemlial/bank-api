package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// "net/url"
	"strconv"
	"time"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

type ApiError struct {
	Error string `json:"error"`
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts", WrapToHandle(s.handle))

	log.Println("Running Server on port: ", s.listenAddr)

	log.Fatal(http.ListenAndServe(s.listenAddr, mux))
}

func (s *APIServer) handle(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGet(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	case "DELETE":
		return s.handleDeleteAccount(w, r)
	case "PUT":
		return s.handleUpdateAccount(w, r)
	}
	return fmt.Errorf("unsupported method: %s", r.Method)
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
	if _, ok := r.URL.Query()["id"]; !ok {
		id, err := getID(r)
		if err != nil {
			return err
		}
		err = s.store.DeleteAccount(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, map[string]int{
			"deleted": id,
		})
	}
	return fmt.Errorf("missing query parameter 'id'")
}

func (s *APIServer) handleGet(w http.ResponseWriter, r *http.Request) error {
	if _, ok := r.URL.Query()["id"]; ok {
		id, err := getID(r)
		if err != nil {
			return err
		}
		return s.handleGetById(w, r, id)
	}
	all, ok := r.URL.Query()["all"]
	if ok && all[0] == "true" {
		return s.handleGetAll(w, r)
	}
	return fmt.Errorf("invalid query")
}

func (s *APIServer) handleGetAll(w http.ResponseWriter, r *http.Request) error {
	accs, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accs)
}

func (s *APIServer) handleGetById(w http.ResponseWriter, r *http.Request, id ...int) error {
	acc, err := s.store.GetAccountByID(id[0])
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, acc)
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

func getID(r *http.Request) (int, error) {
	idStr := r.URL.Query()["id"]
	id, err := strconv.Atoi(idStr[0])
	if err != nil {
		return id, fmt.Errorf("invalid id")
	}
	return id, nil
}
