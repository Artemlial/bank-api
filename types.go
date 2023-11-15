package main

import (
	"io"
	"log"
	"math/rand"
	"time"
)

type Account struct {
	ID                int       `json:"id"`
	Firstname         string    `json:"firstname"`
	Lastname          string    `json:"lastname"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           float64   `json:"balance"`
	CreatedAt         time.Time `json:"created_at"`
}

func NewAccount(firstname, lastname, encPasswd string) *Account {
	return &Account{
		Firstname:         firstname,
		Lastname:          lastname,
		Number:            int64(rand.Intn(1000000)),
		EncryptedPassword: encPasswd,
		CreatedAt:         time.Now().UTC(),
	}
}

type LoginRequest struct {
	AccountNumber int64  `json:"number"`
	Password      string `json:"password"`
}

type CreateAccountRequest struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type Storage interface {
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	CreateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int64) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type TransactionRequest struct {
	DestId int     `json:"dest-id"`
	Amount float64 `json:"amount"`
}

type JWTClaims struct {
	ExpiredAt     int
	AccountNumber int64
	UserType      string
}

func NewJWTClaims(expAt int, accNum int64, userType string) *JWTClaims {
	return &JWTClaims{
		ExpiredAt:     expAt,
		AccountNumber: accNum,
		UserType:      userType,
	}
}

func (j *JWTClaims) Valid() error {
	return nil
}

type MyLogger struct {
	logger *log.Logger
}

func GetMyLogger(output io.Writer, flags int) *MyLogger {
	ml := &MyLogger{}
	ml.logger = log.New(output, "", flags)
	return ml
}

func (l *MyLogger) event(message string) {
	l.logger.SetPrefix("[EVENT]")
	l.logger.Println(message)
}

func (l *MyLogger) susEvent(message string) {
	l.logger.SetPrefix("[SUS EVENT]")
	l.logger.Println(message)
}

func (l *MyLogger) error(message error) {
	l.logger.SetPrefix("[ERR]")
	l.logger.Println(message.Error())
}

func (l *MyLogger) fatal(err error) {
	l.logger.SetPrefix("[FATAL ERR]")
	l.logger.Fatalln(err.Error())
}
