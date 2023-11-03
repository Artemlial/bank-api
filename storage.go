package main

import (
	// inner
	"database/sql"
	"fmt"
	"log"

	// outer
	_ "github.cim/lib/pq" //psql driver
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
	return nil
}

func (p *PsqlStorage) UpdateAccount(acc *Account) error {
	return nil
}

func (p *PsqlStorage) GetAccountByID(id int) (*Account, error) {
	rows,err:=p.db.Query(`select * from Accounts where id = $1`,id)
	if err!=nil{
		return nil,err
	}
	d:=&Account{}
	err=rows.Scan(&d.ID,&d.Firstname,&d.Lastname,&d.Number,&d.Balance,&d.CreatedAt)
	if err!=nil{
		return nil,err
	}
		
	return d, nil
}

func (p *PsqlStorage) GetAccounts() ([]*Account,error){
	var accs []*Account
	rows,err:=p.db.Query(`select * from Accounts`)
	if err!=nil{
		return accs,err
	}

	for rows.Next(){
		d:=&Account{}
		err:=rows.Scan(&d.ID,&d.Firstname,&d.Lastname,&d.Number,&d.Balance,&d.CreatedAt)
		if err!=nil{
			return accs,err
		}
		if err:=rows.Err();err!=nil{
			return accs,err
		}
		accs = append(accs,d)
	}
	return accs,nil
}
