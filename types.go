package main

import (
	"math/rand"
	"time"
)

type Account struct {
	ID        int       `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Number    int64     `json:"number"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(firstname, lastname string) *Account {
	return &Account{
		Firstname: firstname,
		Lastname:  lastname,
		Number:    int64(rand.Intn(1000000)),
		CreatedAt: time.Now().UTC(),
	}
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
	GetAccounts()([]*Account,error)
}
