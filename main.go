package main

import (
	"os"
	yml "gopkg.in/yaml.v2"
)

var cfg Config

func init(){
	f,err:=os.Open("conf.yml")
	if err!=nil{
		panic(err)
	}
	defer f.Close()
	err=yml.NewDecoder(f).Decode(&cfg)
	if err!=nil{
		panic(err)
	}
}

func main() {
	storage, err := NewPsqlStorage(cfg.DB.Host,cfg.DB.Port,cfg.DB.User,cfg.DB.Pswd)
	storage.Init()
	if err != nil {
		panic(err)
	}
	defer storage.db.Close()
	apiserver := NewAPIServer(cfg.Server.Port, storage)
	apiserver.Run()
}
