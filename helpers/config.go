package helpers

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	BindAddress       string  `mapstructure:"bind_address"`
	Port              string  `mapstructure:"listen_port"`
	ProxyProtocolPort string  `mapstructure:"proxyprotocol_port"`
	ServerLat         float64 `mapstructure:"server_lat"`
	ServerLng         float64 `mapstructure:"server_lng"`
	IPInfoAPIKey      string  `mapstructure:"ipinfo_api_key"`

	StatsPassword string `mapstructure:"statistics_password"`
	RedactIP      bool   `mapstructure:"redact_ip_addresses"`

	AssetsPath string `mapstructure:"assets_path"`

	DatabaseType     string `mapstructure:"database_type"`
	DatabaseHostname string `mapstructure:"database_hostname"`
	DatabaseName     string `mapstructure:"database_name"`
	DatabaseUsername string `mapstructure:"database_username"`
	DatabasePassword string `mapstructure:"database_password"`

	DatabaseFile string `mapstructure:"database_file"`
}

var (
	loadedConfig *Config = nil
)

func init() {
	viper.SetDefault("listen_port", "8989")
	viper.SetDefault("proxyprotocol_port", "0")
	viper.SetDefault("download_chunks", 4)
	viper.SetDefault("distance_unit", "K")
	viper.SetDefault("enable_cors", false)
	viper.SetDefault("statistics_password", "PASSWORD")
	viper.SetDefault("redact_ip_addresses", false)
	viper.SetDefault("assets_path", "./assets")
	viper.SetDefault("database_type", "postgresql")
	viper.SetDefault("database_hostname", "localhost")
	viper.SetDefault("database_name", "speedtest")
	viper.SetDefault("database_username", "postgres")

	viper.SetConfigName("settings")
	viper.AddConfigPath(".")
}

func Load() Config {
	var conf Config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("No config file found in search paths, using default values")
		} else {
			log.Fatalf("Error reading config: %s", err)
		}
	}

	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("Error parsing config: %s", err)
	}

	loadedConfig = &conf

	return conf
}

func LoadFile(configFile string) Config {
	var conf Config

	f, err := os.OpenFile(configFile, os.O_RDONLY, 0444)

	if err != nil {
		log.Fatalf("Failed to open config file: %s", err)
	}

	defer f.Close()

	if err := viper.ReadConfig(f); err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("Error parsing config: %s", err)
	}

	loadedConfig = &conf

	return conf
}

func LoadedConfig() *Config {
	if loadedConfig == nil {
		Load()
	}
	return loadedConfig
}
