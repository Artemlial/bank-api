package main

import (
	// inner
	"database/sql"
	"fmt"
	"io"
	"os"

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

func (p *SqlStorage) Init(flag bool, cfg *Config) error {
	if !flag {
		return nil
	}
	return p.CreateSchema(cfg)
}

func (p *SqlStorage) CreateSchema(cfg *Config) error {
	MyLog.event("Changing schema")
	err := p.DropAccounts(cfg.DB.PathToScripts)
	if err != nil {
		MyLog.error(fmt.Errorf("cannot drop table : %s", err.Error()))
		return err
	}

	query, err := GetScript(cfg.DB.PathToScripts + "up.sql")
	if err != nil {
		MyLog.fatal(err)
	}

	_, err = p.db.Query(query)
	if err != nil {
		MyLog.error(err)
	}
	return err
}

func (p *SqlStorage) DropAccounts(path string) error {
	script, err := GetScript(path + "down.sql")
	if err != nil {
		MyLog.fatal(err)
	}
	_, err = p.db.Exec(script)
	return err
}

func (p *SqlStorage) CreateAccount(acc *Account) error {
	_, err := p.db.Exec(`
	createAccount($1,$2,$3,$4,$5,$6)
	`, acc.Firstname, acc.Lastname, acc.Number, acc.EncryptedPassword, acc.Balance, acc.CreatedAt)
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
	updateAccount(%s,%s,%.2f,%d)
	`, acc.Firstname, acc.Lastname, acc.Balance, acc.ID)
	_, err := p.db.Exec(query)
	if err != nil {
		MyLog.error(err)
		return fmt.Errorf("internal server error")
	}
	MyLog.event(fmt.Sprintf("updated account %+v\n", acc))
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

func (p *SqlStorage) GetAccountByNumber(number int64) (*Account, error){
	rows,err:= p.db.Query("SELECT * FROM Accounts WHERE number = $1",number)
	if err!=nil{
		return nil,err
	}
	acc,err := scanRowsIntoAccount(rows)

	if err!=nil{
		return nil,err
	}
	return acc,nil
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

func GetScript(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if b, err := io.ReadAll(f); err == nil {
		return string(b), nil
	}
	return "", err
}
