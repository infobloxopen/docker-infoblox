package common

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/caarlos0/env"
	"os"
	"path"
)

const configFileDir = "/etc/infoblox"

type Config interface {
	LoadFromConfFile() error
	LoadConfig() error
}

type GridConfig struct {
	GridHost            string `toml:"grid-host" env:"GRID_HOST"`
	WapiVer             string `toml:"wapi-version" env:"WAPI_VERSION"`
	WapiPort            string `toml:"wapi-port" env:"WAPI_PORT"`
	WapiUsername        string `toml:"wapi-username" env:"WAPI_USERNAME"`
	WapiPassword        string `toml:"wapi-password" env:"WAPI_PASSWORD"`
	SslVerify           string `toml:"ssl-verify" env:"SSL_VERIFY"`
	HttpRequestTimeout  uint   `toml:"http-request-timeout" env:"HTTP_REQUEST_TIMEOUT"`
	HttpPoolConnections uint   `toml:"http-pool-connections" env:"HTTP_POOL_CONNECTIONS"`
}

type IpamConfig struct {
	GlobalNetview          string `toml:"global-view" env:"GLOBAL_VIEW"`
	GlobalNetworkContainer string `toml:"global-network-container" env:"GLOBAL_NETWORK_CONTAINER"`
	GlobalPrefixLength     uint   `toml:"global-prefix-length" env:"GLOBAL_PREFIX_LENGTH"`
	LocalNetview           string `toml:"local-view" env:"LOCAL_VIEW"`
	LocalNetworkContainer  string `toml:"local-network-container" env:"LOCAL_NETWORK_CONTAINER"`
	LocalPrefixLength      uint   `toml:"local-prefix-length" env:"LOCAL_PREFIX_LENGTH"`
}

type PluginConfig struct {
	ConfigFile string `toml:"" env:"CONF_FILE_NAME"`
	Debug      bool   `toml:"debug" env:"DEBUG"`
	GridConfig `toml:"grid-config"`
	IpamConfig `toml:"ipam-config"`
}

type CreateEADefConfig struct {
	ConfigFile string `toml:""`
	Debug      bool   `toml:"debug"`
	GridConfig `toml:"grid-config"`
}

func NewGridConfig() GridConfig {
	return GridConfig{
		GridHost:            "",
		WapiVer:             "2.0",
		WapiPort:            "443",
		WapiUsername:        "",
		WapiPassword:        "",
		SslVerify:           "false",
		HttpRequestTimeout:  60,
		HttpPoolConnections: 10,
	}
}

func NewPluginConfig() PluginConfig {
	return PluginConfig{
		Debug: false,

		GridConfig: NewGridConfig(),
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

func NewCreateEADefConfig() CreateEADefConfig {
	return CreateEADefConfig{
		Debug:      false,
		GridConfig: NewGridConfig(),
	}
}

func ReadFromConfigFile(filename string, config Config) error {
	// load variables from the config file
	if filename != "" {
		configFilePath := path.Join(configFileDir, filename)
		logrus.Infof("Loading configuration from file %s\n", configFilePath)
		if _, err := toml.DecodeFile(configFilePath, config); err != nil {
			logrus.Errorf("Cannot load the configuration file %s, %s\n", configFilePath, err)
			return err
		}
	}
	return nil
}

func (pc *PluginConfig) LoadFromConfFile() error {
	// Look for CONF_FILE_NAME in the env variables
	if err := pc.LoadFromEnv(); err != nil {
		return err
	}

	return ReadFromConfigFile(pc.ConfigFile, pc)
}

func (pc *PluginConfig) LoadFromEnv() error {
	logrus.Debugln("Loading IPAM Configuration from the environment variables")

	// Load environment variables from the PluginConfig variables other that struct type
	if err := env.Parse(pc); err != nil {
		logrus.Errorf("Failed to parse PluginConfig environment variables, %s", err)
		return err
	}

	// Look for the environment variables in GridConfig struct
	if err := env.Parse(&pc.GridConfig); err != nil {
		logrus.Errorf("Failed to parse GridConfig environment variables, %s", err)
		return err
	}

	// Look for the environment variables in IpamConfig struct
	if err := env.Parse(&pc.IpamConfig); err != nil {
		logrus.Errorf("Failed to parse IpamConfig environment variables, %s", err)
		return err
	}

	return nil
}

func (pc *PluginConfig) LoadConfig() error {
	logrus.Infof("Loading Plugin Configuration")
	if err := pc.LoadFromConfFile(); err != nil {
		return err
	}

	if err := pc.LoadFromEnv(); err != nil {
		return err
	}

	logrus.Infof("Configuration successfully loaded\n")
	return nil
}

func (eac *CreateEADefConfig) LoadFromCommandLine() error {
	// Load configuration from the command line arguments
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagSet.StringVar(&eac.ConfigFile, "conf-file", eac.ConfigFile, "File path of configuration file")
	flagSet.BoolVar(&eac.Debug, "debug", eac.Debug, "Sets log level to debug")
	flagSet.StringVar(&eac.GridHost, "grid-host", eac.GridHost, "IP of Infoblox Grid Host")
	flagSet.StringVar(&eac.WapiVer, "wapi-version", eac.WapiVer, "Infoblox WAPI Version.")
	flagSet.StringVar(&eac.WapiPort, "wapi-port", eac.WapiPort, "Infoblox WAPI Port.")
	flagSet.StringVar(&eac.WapiUsername, "wapi-username", eac.WapiUsername, "Infoblox WAPI Username")
	flagSet.StringVar(&eac.WapiPassword, "wapi-password", eac.WapiPassword, "Infoblox WAPI Password")
	flagSet.StringVar(&eac.SslVerify, "ssl-verify", eac.SslVerify, "Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.")
	flagSet.UintVar(&eac.HttpRequestTimeout, "http-request-timeout", eac.HttpRequestTimeout, "Infoblox WAPI request timeout in seconds.")
	flagSet.UintVar(&eac.HttpPoolConnections, "http-pool-connections", eac.HttpPoolConnections, "Infoblox WAPI connection pool size.")

	flagSet.Parse(os.Args[1:])
	return nil
}

func (eac *CreateEADefConfig) LoadFromConfFile() error {
	// look for --conf-file flag in the cmd line args
	if err := eac.LoadFromCommandLine(); err != nil {
		return err
	}

	return ReadFromConfigFile(eac.ConfigFile, eac)
}

func (eac *CreateEADefConfig) LoadConfig() error {
	logrus.Infof("Loading CreateEaDefs Configuration")
	if err := eac.LoadFromConfFile(); err != nil {
		return err
	}

	if err := eac.LoadFromCommandLine(); err != nil {
		return err
	}

	logrus.Infof("Configuration successfully loaded\n")
	return nil
}

func LoadPluginConfig() (*PluginConfig, error) {
	pc := NewPluginConfig()
	err := pc.LoadConfig()
	return &pc, err
}

func LoadCreateEADefConfig() (*CreateEADefConfig, error) {
	eac := NewCreateEADefConfig()
	err := eac.LoadConfig()
	return &eac, err
}
