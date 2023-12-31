package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	// "net/url"
	"strconv"
	"time"

	// outer
	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
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
	// mux.HandleFunc("/register",WrapToHandle(s.handleRegister))
	mux.HandleFunc("/login", WrapToHandle(s.handleLogin))
	mux.HandleFunc("/api/accounts", WrapToHandle(s.handle))
	mux.HandleFunc("/api/accounts/id", jwtAuthWrapper(WrapToHandle(s.handleID), s.store))

	MyLog.event("starting server on port")

	MyLog.fatal(http.ListenAndServe(s.listenAddr, mux))
}

// this one's for later
// func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) error{
// 	if r.Method!="POST"{
// 		return fmt.Errorf("method %s not supported",r.Method)
// 	}
// 	var req *CreateAccountRequest
// }

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method %s not supported", r.Method)
	}
	req := &LoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		MyLog.error(fmt.Errorf("cannot read request Body as JSON: %s", err.Error()))
		return err
	}
	acc, err := s.store.GetAccountByNumber(req.AccountNumber)
	if err != nil {
		MyLog.error(err)
		return fmt.Errorf("cannot verify your credentials")
	}
	if acc.Number != req.AccountNumber {
		deny(w)
		return nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(acc.EncryptedPassword), []byte(req.Password))
	if err != nil {
		deny(w)
		return nil
	}

	return nil
}

func (s *APIServer) handle(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAll(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	case "PUT":
		return s.handleUpdateAccount(w, r)
	}
	return fmt.Errorf("unsupported method: %s", r.Method)
}

func (s *APIServer) handleID(w http.ResponseWriter, r *http.Request) error {
	if id, err := getID(r); err == nil {
		switch r.Method {
		case "GET":
			return s.handleGetById(w, r, id)
		case "DELETE":
			return s.handleDeleteAccount(w, r, id)
		default:
			return fmt.Errorf("unsupported method")
		}
	}
	return fmt.Errorf("missing query parameter 'id'")
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	acc := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(acc); err != nil {
		MyLog.error(err)
		return fmt.Errorf("cannot parse json in request body")
	}
	account := NewAccount(acc.Firstname, acc.Lastname, "")
	token, err := createJWT(account)
	if err != nil {
		MyLog.error(err)
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, struct {
		AuthToken  string
		Created_at time.Time
	}{
		AuthToken:  token,
		Created_at: account.CreatedAt,
	})
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request, id int) error {
	err := s.store.DeleteAccount(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{
		"deleted": id,
	})
}

func (s *APIServer) handleGetAll(w http.ResponseWriter, r *http.Request) error {
	accs, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accs)
}

func (s *APIServer) handleGetById(w http.ResponseWriter, r *http.Request, id int) error {
	acc, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, acc)
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	acc := &Account{}
	err := json.NewDecoder(r.Body).Decode(acc)
	if err != nil {
		return err
	}
	r.Body.Close()

	return s.store.UpdateAccount(acc)
}

// make error methods later
func jwtAuthWrapper(handlefunc http.HandlerFunc, store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		token, err := validateJWT(tokenString)
		if err != nil {
			MyLog.error(err)
			deny(w)
			return
		}
		if !token.Valid {
			MyLog.error(err)
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}

		id, err := getID(r)
		if err != nil {
			MyLog.error(err)
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid query"})
			return
		}

		acc, err := store.GetAccountByID(id)
		if err != nil {
			deny(w)
			return
		}

		claims, ok := token.Claims.(*JWTClaims)

		if !ok {
			MyLog.error(fmt.Errorf("token.Claims conversion error"))
			WriteJSON(w, http.StatusInternalServerError, ApiError{Error: "Oopsie ;p"})
			return
		}

		if acc.Number != claims.AccountNumber {
			MyLog.susEvent("api.go::Token Data Does Not Match Any Account!!!!!")
			deny(w)
			return
		}

		handlefunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT), nil
	})
}

func createJWT(account *Account) (string, error) {
	claims := NewJWTClaims(15000, account.Number, "undefined")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(cfg.JWT))
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func WrapToHandle(f apiFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("content-type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func getID(r *http.Request) (int, error) {
	idStr, ok := r.URL.Query()["id"]
	if !ok {
		return 0, fmt.Errorf("missing query parameter 'id'")
	}
	id, err := strconv.Atoi(idStr[0])
	if err != nil {
		return id, fmt.Errorf("invalid id")
	}
	return id, nil
}

func deny(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "access denied"})
}
