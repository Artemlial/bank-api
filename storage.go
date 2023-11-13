package main

import (
	// inner
	"database/sql"
	"fmt"

	// outer
	_ "github.com/lib/pq" //psql driver
)

var uriTemplates = map[string]string{
	"postgres": "postgres://%s:%s@%s:%s/?sslmode=disable",
	"mysql":    "%s:%s@tcp(%s:%s)",
}

func GetStorageCredentials(db *Config) (string, string) {
	dbName := cfg.DB.Name
	user := cfg.DB.User
	pswd := cfg.DB.Pswd
	host := cfg.DB.Host
	port := cfg.DB.Port
	uriTemplate := uriTemplates[dbName]
	return dbName, fmt.Sprintf(uriTemplate, user, pswd, host, port)
}

type SqlStorage struct {
	db *sql.DB
}

func NewSqlStorage(name, uri string) (*SqlStorage, error) {
	MyLog.event(fmt.Sprintf("connecting to %s\n", name))
	db, err := sql.Open(name, uri)
	if err != nil {
		MyLog.error(err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		MyLog.error(err)
		return nil, err
	}

	psql := &SqlStorage{db: db}
	return psql, nil
}

func (p *SqlStorage) Init() error {
	return p.CreateTableAccounts()
}

func (p *SqlStorage) CreateTableAccounts() error {
	MyLog.event("creating table accounts")
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
		MyLog.error(err)
	}
	return err
}

func (p *SqlStorage) CreateAccount(acc *Account) error {
	_, err := p.db.Exec(`
	insert into Accounts("first_name","last_name","number","balance","created_at") 
	values($1,$2,$3,$4,$5);
	`, acc.Firstname, acc.Lastname, acc.Number, acc.Balance, acc.CreatedAt)
	if err != nil {
		MyLog.error(err)
		return fmt.Errorf("internal server error")
	}

	MyLog.event(fmt.Sprintf("created account number %d\n", acc.Number))

	return nil
}

func (p *SqlStorage) DeleteAccount(id int) error {
	acc, err := p.GetAccountByID(id)
	if err == nil {
		MyLog.event(fmt.Sprintf("deleting account %+v\n", acc))
	}
	_, err = p.db.Query("delete from Accounts where id = $1", id)
	if err != nil {
		MyLog.error(err)
		return fmt.Errorf("internal server error")
	}
	return nil
}

func (p *SqlStorage) UpdateAccount(acc *Account) error {
	query := fmt.Sprintf(`
	update Accounts
	set first_name = '%s',
	    last_name = '%s',
	    balance = %.2f
	where id = %d;
	`, acc.Firstname, acc.Lastname, acc.Balance, acc.ID)
	_, err := p.db.Exec(query)
	MyLog.event(fmt.Sprintf("updated account %+v\n", acc))
	if err != nil {
		MyLog.error(err)
		return fmt.Errorf("internal server error")
	}
	return nil
}

func (p *SqlStorage) GetAccountByID(id int) (*Account, error) {
	rows, err := p.db.Query(`select * from Accounts where id = $1`, id)
	if err != nil {
		MyLog.error(err)
		return nil, fmt.Errorf("internal server error")
	}

	for rows.Next() {
		return scanRowsIntoAccount(rows)
	}
	MyLog.event(fmt.Sprintf("fetched account with id %d", id))
	return nil, fmt.Errorf("Account with id %d not found", id)
}

func (p *SqlStorage) GetAccounts() ([]*Account, error) {
	var accs []*Account
	rows, err := p.db.Query(`select * from Accounts`)
	if err != nil {
		MyLog.error(err)
		return accs, fmt.Errorf("internal server error")
	}

	for rows.Next() {

		d, err := scanRowsIntoAccount(rows)
		if err != nil {
			MyLog.error(err)
			return accs, fmt.Errorf("internal server error")
		}
		if err := rows.Err(); err != nil {
			MyLog.error(err)
			return accs, fmt.Errorf("internal server error")
		}
		accs = append(accs, d)
	}
	MyLog.event("fetched all accounts")
	return accs, nil
}

func scanRowsIntoAccount(row *sql.Rows) (*Account, error) {
	d := &Account{}
	err := row.Scan(&d.ID, &d.Firstname, &d.Lastname, &d.Number, &d.Balance, &d.CreatedAt)

	return d, err
}
