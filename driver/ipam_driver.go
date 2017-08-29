package main

import (
	"github.com/Sirupsen/logrus"
	apiclient "github.com/docker/engine-api/client"
	ipamPluginSdk "github.com/docker/go-plugins-helpers/ipam"
	"github.com/infobloxopen/docker-infoblox/common"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	ctx "golang.org/x/net/context"
	"os"
	"reflect"
	"strings"
)

const socketAddress = "/run/docker/plugins/infoblox.sock"

func getDockerID() (dockerID string, err error) {
	dockerID = ""
	err = nil
	context := ctx.Background()

	// Default to Docker API Version corresponding to Docker v1.10
	if os.Getenv("DOCKER_API_VERSION") == "" {
		if err = os.Setenv("DOCKER_API_VERSION", "1.22"); err != nil {
			logrus.Infof("Cannot set default Docker API Version: '%s'", err)
			os.Exit(1)
		}
	}
	cli, err := apiclient.NewEnvClient()
	if err != nil {
		return
	}

	inf, err := cli.Info(context)
	if err != nil {
		return
	}
	dockerID = inf.ID
	return dockerID, nil
}

func urlToRequestType(url string) string {
	parts := strings.Split(url, ".")
	n := len(parts)
	if n > 0 {
		n = n - 1
	}

	return parts[n]
}

type ipamCall struct {
	url string
	f   func(r interface{}) (map[string]interface{}, error)
	t   reflect.Type
}

func main() {
	config, err := common.LoadPluginConfig()
	if config == nil || err != nil {
		logrus.Fatal(err)
	}

	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("Socket File: '%s'", socketAddress)
	logrus.Debugf("Configuration options : %+v\n", config)

	hostConfig := ibclient.HostConfig{
		Host:     config.GridHost,
		Version:  config.WapiVer,
		Port:     config.WapiPort,
		Username: config.WapiUsername,
		Password: config.SecuredWapiPassword(),
	}

	transportConfig := ibclient.NewTransportConfig(
		config.SslVerify,
		int(config.HttpRequestTimeout),
		int(config.HttpPoolConnections),
	)

	requestBuilder := &ibclient.WapiRequestBuilder{}
	requestor := &ibclient.WapiHttpRequestor{}

	conn, err := ibclient.NewConnector(hostConfig, transportConfig, requestBuilder, requestor)

	if err != nil {
		logrus.Fatal(err)
	}

	dockerID, err := getDockerID()
	if err != nil {
		logrus.Fatal(err)
	}

	if len(dockerID) > 0 {
		logrus.Infof("Docker id is '%s'\n", dockerID)
	}
	objMgr := ibclient.NewObjectManager(conn, "Docker", dockerID)

	ipamDrv := NewInfobloxDriver(objMgr, config.GlobalNetview, config.GlobalNetworkContainer, config.GlobalPrefixLength,
		config.LocalNetview, config.LocalNetworkContainer, config.LocalPrefixLength)

	h := ipamPluginSdk.NewHandler(ipamDrv)
	h.ServeUnix(socketAddress, 0)
}
