package common

import (
	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/caarlos0/env"
	"path"
)

const configFileDir = "/etc/infoblox"

type GridConfig struct {
	GridHost            string `toml:"grid_host" env:"GRID_HOST"`
	WapiVer             string `toml:"wapi_version" env:"WAPI_VERSION"`
	WapiPort            string `toml:"wapi_port" env:"WAPI_PORT"`
	WapiUsername        string `toml:"wapi_username" env:"WAPI_USERNAME"`
	WapiPassword        string `toml:"wapi_password" env:"WAPI_PASSWORD"`
	SslVerify           string `toml:"ssl_verify" env:"SSL_VERIFY"`
	HttpRequestTimeout  uint   `toml:"http_request_timeout" env:"HTTP_REQUEST_TIMEOUT"`
	HttpPoolConnections uint   `toml:"http_pool_connections" env:"HTTP_POOL_CONNECTIONS"`
}

type IpamConfig struct {
	GlobalNetview          string `toml:"global_view" env:"GLOBAL_VIEW"`
	GlobalNetworkContainer string `toml:"global_network_container" env:"GLOBAL_NETWORK_CONTAINER"`
	GlobalPrefixLength     uint   `toml:"global_prefix_length" env:"GLOBAL_PREFIX_LENGTH"`
	LocalNetview           string `toml:"local_view" env:"LOCAL_VIEW"`
	LocalNetworkContainer  string `toml:"local_network_container" env:"LOCAL_NETWORK_CONTAINER"`
	LocalPrefixLength      uint   `toml:"local_prefix_length" env:"LOCAL_PREFIX_LENGTH"`
}

type Config struct {
	ConfigFile string `toml:"" env:"CONF_FILE_NAME"`
	Debug      bool   `toml:"debug" env:"DEBUG"`
	GridConfig `toml:"grid_config"`
	IpamConfig `toml:"ipam_config"`
}

func NewConfig() Config {
	return Config{
		ConfigFile: "",
		Debug:      false,

		GridConfig: GridConfig{
			GridHost:            "",
			WapiVer:             "2.0",
			WapiPort:            "443",
			WapiUsername:        "",
			WapiPassword:        "",
			SslVerify:           "false",
			HttpRequestTimeout:  60,
			HttpPoolConnections: 10,
		},

		IpamConfig: IpamConfig{
			GlobalNetview:          "default",
			GlobalNetworkContainer: "",
			GlobalPrefixLength:     24,
			LocalNetview:           "default",
			LocalNetworkContainer:  "",
			LocalPrefixLength:      24,
		},
	}
}

func LoadFromConfFile(config *Config) error {
	// Look for the CONF_FILE_NAME environment variable
	// and load in the Config struct
	logrus.Infoln("Loading IPAM Configuration from the file")

	err := env.Parse(config)
	if err != nil {
		return err
	}

	// Now load config file
	if config.ConfigFile != "" {
		configFilePath := path.Join(configFileDir, config.ConfigFile)
		logrus.Infof("Found Configuration file %s\n", configFilePath)
		if _, err = toml.DecodeFile(configFilePath, config); err != nil {
			logrus.Errorf("Cannot load the configuration file %s, %s\n", configFilePath, err)
			return err
		}
	}
	return nil
}

func LoadConfig() (*Config, error) {
	config := NewConfig()

	if err := LoadFromConfFile(&config); err != nil {
		return &config, err
	}

	logrus.Infoln("Loading IPAM Configuration from the environment variables")
	// Look for the environment variables in GridConfig struct
	if err := env.Parse(&config.GridConfig); err != nil {
		logrus.Errorf("Failed to parse GridConfig environment variables, %s", err)
		return &config, err
	}

	// Look for the environment variables in IpamConfig struct
	if err := env.Parse(&config.IpamConfig); err != nil {
		logrus.Errorf("Failed to parse IpamConfig environment variables, %s", err)
		return &config, err
	}

	logrus.Infof("Configuration successfully loaded\n")
	return &config, nil
}
