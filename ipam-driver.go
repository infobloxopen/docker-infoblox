package main

import (
	"encoding/json"
	"fmt"
	apiclient "github.com/moby/moby/client"
	ipamsapi "github.com/docker/libnetwork/ipams/remote/api"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	ctx "golang.org/x/net/context"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
)

func getDockerID() (dockerID string, err error) {
	dockerID = ""
	err = nil
	context := ctx.Background()

	// Default to Docker API Version corresponding to Docker v1.10
	if os.Getenv("DOCKER_API_VERSION") == "" {
		if err = os.Setenv("DOCKER_API_VERSION", "1.22"); err != nil {
			log.Panicf("Cannot set default Docker API Version: '%s'", err)
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

	return
}

func dirExists(dirname string) (bool, error) {
	fileInfo, err := os.Stat(dirname)
	if err == nil {
		if fileInfo.IsDir() {
			return true, nil
		} else {
			return false, nil
		}
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createDir(dirname string) error {
	return os.MkdirAll(dirname, 0700)
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)

	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

func setupSocket(pluginDir string, driverName string) string {
	exists, err := dirExists(pluginDir)
	if err != nil {
		log.Panicf("Stat Plugin Directory error '%s'", err)
		os.Exit(1)
	}
	if !exists {
		err = createDir(pluginDir)
		if err != nil {
			log.Panicf("Create Plugin Directory error: '%s'", err)
			os.Exit(1)
		}
		log.Printf("Created Plugin Directory: '%s'", pluginDir)
	}

	socketFile := pluginDir + "/" + driverName + ".sock"
	exists, err = fileExists(socketFile)
	if err != nil {
		log.Panicf("Stat Socket File error: '%s'", err)
		os.Exit(1)
	}
	if exists {
		err = deleteFile(socketFile)
		if err != nil {
			log.Panicf("Delete Socket File error: '%s'", err)
			os.Exit(1)
		}
		log.Printf("Deleted Old Socket File: '%s'", socketFile)
	}

	return socketFile
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
	config, err := LoadConfig()
	if config == nil || err != nil {
		log.Fatal(err)
	}

	socketFile := setupSocket(config.PluginDir, config.DriverName)
	log.Printf("Driver Name: '%s'", config.DriverName)
	log.Printf("Socket File: '%s'", socketFile)

	hostConfig := ibclient.HostConfig{
		Host:     config.GridHost,
		Version:  config.WapiVer,
		Port:     config.WapiPort,
		Username: config.WapiUsername,
		Password: config.WapiPassword,
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
		log.Fatal(err)
	}

	dockerID, err := getDockerID()
	if err != nil {
		log.Fatal(err)
	}

	if len(dockerID) > 0 {
		log.Printf("Docker id is '%s'\n", dockerID)
	}
	objMgr := ibclient.NewObjectManager(conn, "Docker", dockerID)

	ipamDrv := NewInfobloxDriver(objMgr, config.GlobalNetview, config.GlobalNetworkContainer, config.GlobalPrefixLength,
		config.LocalNetview, config.LocalNetworkContainer, config.LocalPrefixLength)
	ipamCalls := []ipamCall{
		{"/Plugin.Activate", ipamDrv.PluginActivate, nil},
		{"/IpamDriver.GetCapabilities", ipamDrv.GetCapabilities, nil},
		{"/IpamDriver.GetDefaultAddressSpaces", ipamDrv.GetDefaultAddressSpaces, nil},
		{"/IpamDriver.RequestPool", ipamDrv.RequestPool,
			reflect.TypeOf(ipamsapi.RequestPoolRequest{})},
		{"/IpamDriver.ReleasePool", ipamDrv.ReleasePool,
			reflect.TypeOf(ipamsapi.ReleasePoolRequest{})},
		{"/IpamDriver.RequestAddress", ipamDrv.RequestAddress,
			reflect.TypeOf(ipamsapi.RequestAddressRequest{})},
		{"/IpamDriver.ReleaseAddress", ipamDrv.ReleaseAddress,
			reflect.TypeOf(ipamsapi.ReleaseAddressRequest{})},
	}

	handlers := make(map[string]ipamCall)

	for _, v := range ipamCalls {
		handlers[v.url] = v
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		requestType := urlToRequestType(url)

		if c, ok := handlers[url]; ok {

			//var req interface{}
			var req interface{}
			if c.t != nil {
				req = reflect.New(c.t).Interface()
				if err := json.NewDecoder(r.Body).Decode(req); err != nil {
					log.Printf("%s: Bad Request Error: %s\n", url, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			log.Printf("'%s' request: '%s'\n", requestType, req)
			res, err := c.f(req)
			if err != nil || res == nil {
				if res == nil {
					log.Printf("IPAM Driver returned nil result")
					res = make(map[string]interface{})
				}
				if err != nil {
					log.Printf("IPAM Driver error '%s'", err)
					res["Error"] = err.Error()
				}
			}
			log.Printf("'%s' result: '%s'\n", requestType, res)

			if err := json.NewEncoder(w).Encode(res); err != nil {
				log.Printf("%s: Bad Response Error: %s\n", url, err)
			}
		}
		fmt.Fprintf(w, "{ \"Error\": \"%s\"}", url)
	})

	l, err := net.Listen("unix", socketFile)
	if err != nil {
		log.Panic(err)
	}
	if err := http.Serve(l, nil); err != nil {
		log.Panic(err)
	}

	os.Exit(0)
}
