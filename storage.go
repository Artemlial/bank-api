package main

import (
	// inner
	"database/sql"
	"fmt"
	"log"

	// outer
	_ "github.com/lib/pq" //psql driver
)

type PsqlStorage struct {
	db *sql.DB
}

func NewPsqlStorage() (*PsqlStorage, error) {
	db, err := sql.Open("postgres", "postgres://postgres:p@55w0rdPG!/localhost:3000/?sslmode=disable") //get rid of hardcode and store pg info in .config
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	psql := &PsqlStorage{db: db}
	return psql, nil
}

func (p *PsqlStorage) Init() error {
	return p.CreateTableAccounts()
}

func (p *PsqlStorage) CreateTableAccounts() error {
	query := `create table if not exists Accounts(
		id serial primary key ,
		first_name varchar(20),
		last_name varchar(20),
		number bigint,
		balance money,
		created_at timestamptz);`

	_, err := p.db.Query(query)
	return err
}

func (p *PsqlStorage) CreateAccount(acc *Account) error {
	query := fmt.Sprintf(`
	insert into Accounts("first_name","last_name","number","created_at") 
	values(%s,%s,%d,%v)
	`, acc.Firstname, acc.Lastname, acc.Number, acc.CreatedAt)

	resp, err := p.db.Exec(query)

	if err != nil {
		return err
	}

	log.Printf("CreateAccount Psql response: %+v\n", resp)

	return nil
}

func (p *PsqlStorage) DeleteAccount(id int) error {
	_, err := p.db.Query("delete from account where id = $1", id)
	return err
}

func (p *PsqlStorage) UpdateAccount(acc *Account) error {
	return nil
}

func (p *PsqlStorage) GetAccountByID(id int) (*Account, error) {
	rows, err := p.db.Query(`select * from Accounts where id = $1`, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanRowsIntoAccount(rows)
	}

	return nil, fmt.Errorf("Account with id %d not found", id)
}

func (p *PsqlStorage) GetAccounts() ([]*Account, error) {
	var accs []*Account
	rows, err := p.db.Query(`select * from Accounts`)
	if err != nil {
		return accs, err
	}

	for rows.Next() {

		d, err := scanRowsIntoAccount(rows)
		if err != nil {
			return accs, err
		}
		if err := rows.Err(); err != nil {
			return accs, err
		}
		accs = append(accs, d)
	}
	return accs, nil
}

func scanRowsIntoAccount(row *sql.Rows) (*Account, error) {
	d := &Account{}
	err := row.Scan(&d.ID, &d.Firstname, &d.Lastname, &d.Number, &d.Balance, &d.CreatedAt)

	return d, err
}
