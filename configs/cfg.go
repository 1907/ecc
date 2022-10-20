package configs

var C Config

type Config struct {
	Env   string `yaml:"env"`
	Mongo struct {
		DNS        string `yaml:"dns"`
		Db         string `yaml:"db"`
		Collection string `yaml:"collection"`
	} `yaml:"mongo"`
	Mysql struct {
		Alias string `yaml:"alias"`
		Dns   string `yaml:"dns"`
	} `yaml:"mysql"`
}
