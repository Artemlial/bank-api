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

func NewPsqlStorage(host, port, user, pswd string) (*PsqlStorage, error) {
	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/?sslmode=disable", user, pswd, host, port)
	db, err := sql.Open("postgres", uri)
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
		created_at timestamptz);
		`

	_, err := p.db.Query(query)
	return err
}

func (p *PsqlStorage) CreateAccount(acc *Account) error {
	resp, err := p.db.Exec(`
	insert into Accounts("first_name","last_name","number","balance","created_at") 
	values($1,$2,$3,$4,$5)
	`, acc.Firstname, acc.Lastname, acc.Number, acc.Balance, acc.CreatedAt)

	if err != nil {
		return err
	}

	log.Printf("CreateAccount Psql response: %+v\n", resp)

	return nil
}

func (p *PsqlStorage) DeleteAccount(id int) error {
	_, err := p.db.Query("delete from Accounts where id = $1", id)
	return err
}

func (p *PsqlStorage) UpdateAccount(acc *Account) error {
	query := fmt.Sprintf(`
	update Accounts
	set first_name = '%s',
	    last_name = '%s',
	    balance = '%s'
	where id = %d;
	`, acc.Firstname, acc.Lastname, acc.Balance, acc.ID)
	res, err := p.db.Exec(query)
	log.Printf("update result %+v", res)
	return err
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
