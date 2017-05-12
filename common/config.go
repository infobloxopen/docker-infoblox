package common

import (
	"flag"
	"github.com/BurntSushi/toml"
	"os"
)

type PluginConfig struct {
	PluginDir  string `toml:"plugin_dir"`
	DriverName string `toml:"driver_name"`
}

type GridConfig struct {
	GridHost            string `toml:"grid_host"`
	WapiVer             string `toml:"wapi_version"`
	WapiPort            string `toml:"wapi_port"`
	WapiUsername        string `toml:"wapi_username"`
	WapiPassword        string `toml:"wapi_password"`
	SslVerify           string `toml:"ssl_verify"`
	HttpRequestTimeout  uint   `toml:"http_request_timeout"`
	HttpPoolConnections uint   `toml:"http_pool_connections"`
}

type IpamConfig struct {
	GlobalNetview          string `toml:"global_view"`
	GlobalNetworkContainer string `toml:"global_container"`
	GlobalPrefixLength     uint   `toml:"global_prefix"`
	LocalNetview           string `toml:"local_view"`
	LocalNetworkContainer  string `toml:"local_container"`
	LocalPrefixLength      uint   `toml:"local_prefix"`
}

type Config struct {
	ConfigFile   string `toml:`
	PluginConfig `toml:"plugin_config"`
	GridConfig   `toml:"grid_config"`
	IpamConfig   `toml:"ipam_config"`
}

func NewConfig() *Config {
	return &Config{
		ConfigFile: "",

		PluginConfig: PluginConfig{
			PluginDir:  "/run/docker/plugins",
			DriverName: "infoblox",
		},

		GridConfig: GridConfig{
			GridHost:            "192.168.124.200",
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
			GlobalNetworkContainer: "global-network-container",
			GlobalPrefixLength:     24,
			LocalNetview:           "default",
			LocalNetworkContainer:  "192.168.0.0/16",
			LocalPrefixLength:      24,
		},
	}
}

func LoadFromCommandLine(config *Config) (*Config, error) {
	if config == nil {
		config = NewConfig()
	}

	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flagSet.StringVar(&config.ConfigFile, "conf-file", "", "File path of configuration file")

	flagSet.StringVar(&config.PluginDir, "plugin-dir", config.PluginDir, "Docker plugin directory where driver socket is created")
	flagSet.StringVar(&config.DriverName, "driver-name", config.DriverName, "Name of Infoblox IPAM driver")

	flagSet.StringVar(&config.GridHost, "grid-host", config.GridHost, "IP of Infoblox Grid Host")
	flagSet.StringVar(&config.WapiVer, "wapi-version", config.WapiVer, "Infoblox WAPI Version.")
	flagSet.StringVar(&config.WapiPort, "wapi-port", config.WapiPort, "Infoblox WAPI Port.")
	flagSet.StringVar(&config.WapiUsername, "wapi-username", config.WapiUsername, "Infoblox WAPI Username")
	flagSet.StringVar(&config.WapiPassword, "wapi-password", config.WapiPassword, "Infoblox WAPI Password")
	flagSet.StringVar(&config.SslVerify, "ssl-verify", config.SslVerify, "Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.")
	flagSet.UintVar(&config.HttpRequestTimeout, "http-request-timeout", config.HttpRequestTimeout, "Infoblox WAPI request timeout in seconds.")
	flagSet.UintVar(&config.HttpPoolConnections, "http-pool-connections", config.HttpPoolConnections, "Infoblox WAPI connection pool size.")

	flagSet.StringVar(&config.GlobalNetview, "global-view", config.GlobalNetview, "Infoblox Network View for Global Address Space")
	flagSet.StringVar(&config.GlobalNetworkContainer, "global-network-container", config.GlobalNetworkContainer, "Subnets will be allocated from this container when --subnet is not specified during network creation")
	flagSet.UintVar(&config.GlobalPrefixLength, "global-prefix-length", config.GlobalPrefixLength, "The default CIDR prefix length when allocating a global subnet.")
	flagSet.StringVar(&config.LocalNetview, "local-view", config.LocalNetview, "Infoblox Network View for Local Address Space")
	flagSet.StringVar(&config.LocalNetworkContainer, "local-network-container", config.LocalNetworkContainer, "Subnets will be allocated from this container when --subnet is not specified during network creation")
	flagSet.UintVar(&config.LocalPrefixLength, "local-prefix-length", config.LocalPrefixLength, "The default CIDR prefix length when allocating a local subnet.")

	flagSet.Parse(os.Args[1:])

	return config, nil
}

func LoadFromConfFile(config *Config) (*Config, error) {
	// Just look for --conf-file flag
	tmpConfig, err := LoadFromCommandLine(NewConfig())
	if tmpConfig == nil || err != nil {
		return tmpConfig, err
	}

	// Now load config file
	if tmpConfig.ConfigFile != "" {
		if _, err = toml.DecodeFile(tmpConfig.ConfigFile, config); err != nil {
			return nil, err
		}
	}

	return config, err

}

func LoadConfig() (config *Config, err error) {
	config = NewConfig()

	if config, err = LoadFromConfFile(config); config == nil || err != nil {
		return
	}

	if config, err = LoadFromCommandLine(config); config == nil || err != nil {
		return
	}

	return
}
