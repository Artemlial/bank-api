package main 


func main() {
	storage,err := NewPsqlStorage()
	if err!=nil{
		panic(err)
	}
	defer storage.db.Close()
	apiserver:=NewAPIServer("",storage)
	apiserver.Run()
}
