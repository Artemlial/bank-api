package main

import (
	"flag"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

var cfg *Config

var MyLog *MyLogger

var changeSchema *bool

func init() {

	changeSchema = flag.Bool("s", false, "s is passed in case you need to change table schema")

	f, err := os.Open("conf.yml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	cfg = &Config{}
	err = yaml.NewDecoder(f).Decode(cfg)
	if err != nil {
		panic(err)
	}
	erF, err := os.OpenFile("./logs/journal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	MyLog = GetMyLogger(erF, log.Ldate|log.Llongfile|log.Ltime)
}

func main() {
	storage, err := NewSqlStorage(GetStorageCredentials(cfg))
	storage.Init(*changeSchema, cfg)
	if err != nil {
		panic(err)
	}
	defer storage.db.Close()
	apiserver := NewAPIServer(cfg.Server.Port, storage)
	apiserver.Run()
}
