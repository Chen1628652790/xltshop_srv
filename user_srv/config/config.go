package config

type MySQLConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	DbName   string `mapstructure:"db_name" json:"db_name"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type ServerConfig struct {
	Mode        string      `mapstructure:"mode" json:"mode"`
	MySQLConfig MySQLConfig `mapstructure:"mysql" json:"mysql"`
}
