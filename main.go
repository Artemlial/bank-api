package main

import (
	"log"
	"os"

	yml "gopkg.in/yaml.v2"
)

var cfg Config

var ErrLog *log.Logger
var GenLog *log.Logger
var SusLog *log.Logger

func init() {
	f, err := os.Open("conf.yml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = yml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		panic(err)
	}
	erF, err := os.OpenFile("./logs/err.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	ErrLog = log.New(erF, "[ERROR]:", log.Ldate|log.Ltime|log.Lshortfile)
	genF, err := os.OpenFile("./logs/general.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	GenLog = log.New(genF, "[EVENT]:", log.Ldate|log.Ltime|log.Lshortfile)
	susF, err := os.OpenFile("./logs/sus.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	SusLog = log.New(susF, "[SUSPICIOUS EVENT]:", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	storage, err := NewPsqlStorage(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Pswd)
	storage.Init()
	if err != nil {
		panic(err)
	}
	defer storage.db.Close()
	apiserver := NewAPIServer(cfg.Server.Port, storage)
	apiserver.Run()
}
