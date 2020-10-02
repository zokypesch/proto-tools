package config

import "github.com/kelseyhightower/envconfig"

// Config struct of configuration
type Config struct {
	// DB Information
	DBAddress  string `envconfig:"DBADDRESS" default:"localhost"`
	DBName     string `envconfig:"DBNAME" default:"proto-db"`
	DBUser     string `envconfig:"DBUSER" default:"root"`
	DBPassword string `envconfig:"DBPASSWORD" default:""`
	DBPort     int    `envconfig:"DBPORT" default:"3306"`
	LogMode    bool   `envconfig:"LOG_MODE" default:"false"`
	// DB Information ETL
	DBAddressETL  string `envconfig:"DBADDRESS" default:"localhost"`
	DBNameETL     string `envconfig:"DBNAME" default:"master_etl"`
	DBNamePipe    string `envconfig:"DBNAME" default:"etl"`
	DBUserETL     string `envconfig:"DBUSER" default:"root"`
	DBPasswordETL string `envconfig:"DBPASSWORD" default:""`
}

// singleton of data
var data *Config

// Get configuration of data
func Get() *Config {
	if data == nil {
		data = &Config{}
		envconfig.MustProcess("", data)
	}

	// returing configuration
	return data
}
