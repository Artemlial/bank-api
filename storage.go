package main

import (
	// inner
	"database/sql"
	"fmt"

	// outer
	_ "github.com/lib/pq" //psql driver
)

type PsqlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(host, port, user, pswd string) (*PsqlStorage, error) {
	GenLog.Println("connecting to postgres")
	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/?sslmode=disable", user, pswd, host, port)
	db, err := sql.Open("postgres", uri)
	if err != nil {
		ErrLog.Println(err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		ErrLog.Println(err.Error())
		return nil, err
	}

	psql := &PsqlStorage{db: db}
	return psql, nil
}

func (p *PsqlStorage) Init() error {
	return p.CreateTableAccounts()
}

func (p *PsqlStorage) CreateTableAccounts() error {
	GenLog.Println("creating table accounts")
	query := `create table if not exists Accounts(
		id serial primary key ,
		first_name varchar(20),
		last_name varchar(20),
		number bigint,
		balance float8,
		created_at timestamptz);
		`

	_, err := p.db.Query(query)
	if err != nil {
		ErrLog.Println(err.Error())
	}
	return err
}

func (p *PsqlStorage) CreateAccount(acc *Account) error {
	_, err := p.db.Exec(`
	insert into Accounts("first_name","last_name","number","balance","created_at") 
	values($1,$2,$3,$4,$5)
	`, acc.Firstname, acc.Lastname, acc.Number, acc.Balance, acc.CreatedAt)

	if err != nil {
		ErrLog.Println(err.Error())
		return fmt.Errorf("internal server error")
	}

	GenLog.Printf("created account number %d\n", acc.Number)

	return nil
}

func (p *PsqlStorage) DeleteAccount(id int) error {
	acc, err := p.GetAccountByID(id)
	if err == nil {
		GenLog.Printf("deleting account %+v\n", acc)
	}
	_, err = p.db.Query("delete from Accounts where id = $1", id)
	if err != nil {
		ErrLog.Println(err.Error())
		return fmt.Errorf("internal server error")
	}
	return nil
}

func (p *PsqlStorage) UpdateAccount(acc *Account) error {
	query := fmt.Sprintf(`
	update Accounts
	set first_name = '%s',
	    last_name = '%s',
	    balance = %.2f
	where id = %d;
	`, acc.Firstname, acc.Lastname, acc.Balance, acc.ID)
	_, err := p.db.Exec(query)
	GenLog.Printf("updated account %+v\n", acc)
	if err != nil {
		ErrLog.Println(err.Error())
		return fmt.Errorf("internal server error")
	}
	return nil
}

func (p *PsqlStorage) GetAccountByID(id int) (*Account, error) {
	rows, err := p.db.Query(`select * from Accounts where id = $1`, id)
	if err != nil {
		ErrLog.Println(err.Error())
		return nil, fmt.Errorf("internal server error")
	}

	for rows.Next() {
		return scanRowsIntoAccount(rows)
	}
	GenLog.Printf("fetched account with id %d",id)
	return nil, fmt.Errorf("Account with id %d not found", id)
}

func (p *PsqlStorage) GetAccounts() ([]*Account, error) {
	var accs []*Account
	rows, err := p.db.Query(`select * from Accounts`)
	if err != nil {
		ErrLog.Println(err.Error())
		return accs, fmt.Errorf("internal server error")
	}

	for rows.Next() {

		d, err := scanRowsIntoAccount(rows)
		if err != nil {
			ErrLog.Println(err.Error())
			return accs, fmt.Errorf("internal server error")
		}
		if err := rows.Err(); err != nil {
			ErrLog.Println(err.Error())
			return accs, fmt.Errorf("internal server error")
		}
		accs = append(accs, d)
	}
	GenLog.Println("fetched all accounts")
	return accs, nil
}

func scanRowsIntoAccount(row *sql.Rows) (*Account, error) {
	d := &Account{}
	err := row.Scan(&d.ID, &d.Firstname, &d.Lastname, &d.Number, &d.Balance, &d.CreatedAt)

	return d, err
}
