package main

import (
	"encoding/json"
	"fmt"
	apitypes "github.com/docker/engine-api/types"
	ipamsapi "github.com/docker/libnetwork/ipams/remote/api"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
)

func dockerApiConnector(proto, addr string) (conn net.Conn, err error) {
	return net.Dial("unix", "/var/run/docker.sock")
}

func getDockerID() (dockerID string, err error) {
	dockerID = ""
	err = nil

	tr := &http.Transport{
		Dial: dockerApiConnector,
	}
	client := &http.Client{Transport: tr}

	var req *http.Request
	req, err = http.NewRequest("GET", "http://fakehost/info", nil)
	if err != nil {
		log.Printf("Cannot create HTTP request: '%s'\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Bad response querying for docker ID: '%s'\n", err)
		return
	}
	if err != nil {
		log.Printf("Error querying for docker ID: '%s'\n", err)
		return
	} else {
		defer resp.Body.Close()
		var contents []byte
		contents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Http Reponse ioutil.ReadAll() Error: '%s'", err)
			return
		}

		apiInfo := new(apitypes.Info)
		err = json.Unmarshal(contents, &apiInfo)

		if err != nil {
			log.Printf("Error unmarshaling docker ID\n: '%s'", apiInfo.ID)
			return
		}
		dockerID = apiInfo.ID
	}

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

type ipamCall struct {
	url string
	f   func(r interface{}) (map[string]interface{}, error)
	t   reflect.Type
}

func main() {
	config := LoadConfig()

	socketFile := setupSocket(config.PluginDir, config.DriverName)
	log.Printf("Driver Name: '%s'", config.DriverName)
	log.Printf("Socket File: '%s'", socketFile)

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

	dockerID, _ := getDockerID()
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
		log.Printf("Plugin: %s\n", url)
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

			res, err := c.f(req)
			if err != nil || res == nil {
				if err != nil {
					log.Printf("IPAM Driver error '%s'", err)
				} else if res == nil {
					log.Printf("IPAM Driver returned nil result")
				}
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			} else {
				if err := json.NewEncoder(w).Encode(res); err != nil {
					log.Printf("%s: Bad Response Error: %s\n", url, err)
				}
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
