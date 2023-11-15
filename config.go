package main

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	DB struct {
		Name          string `yaml:"name"`
		User          string `yaml:"user"`
		Pswd          string `yaml:"pswd"`
		Host          string `yaml:"host"`
		Port          string `yaml:"port"`
		PathToScripts string `yaml:"path_to_scripts"`
	} `yaml:"db"`
	JWT string `yaml:"jwt_key"`
}

// comment for commit
