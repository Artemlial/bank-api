package main

func main() {
	storage, err := NewPsqlStorage()
	storage.Init()
	if err != nil {
		panic(err)
	}
	defer storage.db.Close()
	apiserver := NewAPIServer(":8080", storage)
	apiserver.Run()
}
