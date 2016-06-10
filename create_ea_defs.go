package main

import (
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"log"
)

func main() {
	config := LoadConfig()

	conn, err := ibclient.NewConnector(
		config.GridHost,
		config.WapiVer,
		config.WapiPort,
		config.WapiUsername,
		config.WapiPassword,
		config.SslVerify,
		config.HttpRequestTimeout,
		config.HttpPoolConnections,
		config.HttpPoolMaxSize)

	if err != nil {
		log.Fatal(err)
	}

	objMgr := ibclient.NewObjectManager(conn, "Docker", "")

	reqEaDefs := GetRequiredEADefs()
	for _, e := range reqEaDefs {
		eadef, err := objMgr.GetEADefinition(e.Name)

		if err != nil {
			log.Printf("GetEADefinition(%s) error '%s'", e.Name, err)
			continue
		}

		if eadef != nil {
			log.Printf("EA Definition '%s' already exists", eadef.Name)

		} else {
			log.Printf("EA Definition '%s' not found.", e.Name)
			newEadef, err := objMgr.CreateEADefinition(e)
			if err == nil {
				log.Printf("EA Definition '%s' created", newEadef.Name)
			}
		}
	}
}
