package main

import (
	"flag"
)

type GridConfig struct {
	GridHost     string
	WapiVer      string
	WapiPort     string
	WapiUsername string
	WapiPassword string
	SslVerify    string
}

type DriverConfig struct {
	PluginDir              string
	DriverName             string
	GlobalNetview          string
	GlobalNetworkContainer string
	GlobalPrefixLength     uint
	LocalNetview           string
	LocalNetworkContainer  string
	LocalPrefixLength      uint
}

type Config struct {
	GridConfig
	DriverConfig
}

func LoadConfig() (config *Config) {
	config = new(Config)

	flag.StringVar(&config.GridHost, "grid-host", "192.168.124.200", "IP of Infoblox Grid Host")
	flag.StringVar(&config.WapiVer, "wapi-version", "2.0", "Infoblox WAPI Version.")
	flag.StringVar(&config.WapiPort, "wapi-port", "443", "Infoblox WAPI Port.")
	flag.StringVar(&config.WapiUsername, "wapi-username", "", "Infoblox WAPI Username")
	flag.StringVar(&config.WapiPassword, "wapi-password", "", "Infoblox WAPI Password")
	flag.StringVar(&config.SslVerify, "ssl-verify", "false", "Specifies whether (true/false) to verify server certificate. If a file path is specified, it is assumed to be a certificate file and will be used to verify server certificate.")
	flag.StringVar(&config.PluginDir, "plugin-dir", "/run/docker/plugins", "Docker plugin directory where driver socket is created")
	flag.StringVar(&config.DriverName, "driver-name", "mddi", "Name of Infoblox IPAM driver")
	flag.StringVar(&config.GlobalNetview, "global-view", "default", "Infoblox Network View for Global Address Space")
	flag.StringVar(&config.GlobalNetworkContainer, "global-network-container", "172.18.0.0/16", "Subnets will be allocated from this container when --subnet is not specified during network creation")
	flag.UintVar(&config.GlobalPrefixLength, "global-prefix-length", 24, "The default CIDR prefix length when allocating a global subnet.")
	flag.StringVar(&config.LocalNetview, "local-view", "default", "Infoblox Network View for Local Address Space")
	flag.StringVar(&config.LocalNetworkContainer, "local-network-container", "192.168.0.0/16", "Subnets will be allocated from this container when --subnet is not specified during network creation")
	flag.UintVar(&config.LocalPrefixLength, "local-prefix-length", 24, "The default CIDR prefix length when allocating a local subnet.")

	flag.Parse()

	return config
}
