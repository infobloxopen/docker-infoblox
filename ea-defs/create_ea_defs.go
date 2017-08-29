package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/infobloxopen/docker-infoblox/common"
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

func GetRequiredEADefs() []ibclient.EADefinition {
	ea_defs := common.RequiredEADefs
	res := make([]ibclient.EADefinition, len(ea_defs))
	for i, d := range ea_defs {
		res[i] = *ibclient.NewEADefinition(d)
	}

	return res
}

func main() {
	config, err := common.LoadCreateEADefConfig()
	if config == nil || err != nil {
		logrus.Fatal(err)
	}

	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

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

	objMgr := ibclient.NewObjectManager(conn, "Docker", "")

	reqEaDefs := GetRequiredEADefs()
	for _, e := range reqEaDefs {
		eadef, err := objMgr.GetEADefinition(e.Name)

		if err != nil {
			logrus.Printf("GetEADefinition(%s) error '%s'", e.Name, err)
			continue
		}

		if eadef != nil {
			logrus.Printf("EA Definition '%s' already exists", eadef.Name)

		} else {
			logrus.Printf("EA Definition '%s' not found.", e.Name)
			newEadef, err := objMgr.CreateEADefinition(e)
			if err == nil {
				logrus.Printf("EA Definition '%s' created", newEadef.Name)
			}
		}
	}
}
