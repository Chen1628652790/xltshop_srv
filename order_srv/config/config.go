package config

type MySQLConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	DbName   string `mapstructure:"db_name" json:"db_name"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type ServerConfig struct {
	ServerName            string             `mapstructure:"server_name" json:"server_name"`
	Tags                  []string           `mapstructure:"tags" json:"tags"`
	Host                  string             `mapstructure:"host" json:"host"`
	Mode                  string             `mapstructure:"mode" json:"mode"`
	MySQLConfig           MySQLConfig        `mapstructure:"mysql" json:"mysql"`
	ConsulConfig          ConsulConfig       `mapstructure:"consul" json:"consul"`
	GoodsServerConfig     GoodsSrvConfig     `mapstructure:"goods_srv" json:"goods_srv"`
	InventoryServerConfig InventorySrvConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host" json:"host"`
	Port      int    `mapstructure:"port" json:"port"`
	Namespace string `mapstructure:"namespace" json:"namespace"`
	User      string `mapstructure:"user" json:"user"`
	Password  string `mapstructure:"password" json:"password"`
	DataID    string `mapstructure:"data_id" json:"data_id"`
	Group     string `mapstructure:"group" json:"group"`
}

type GoodsSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type InventorySrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}
