package main

import (
	"encoding/json"
	"flag"
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

		return
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

var np string
var objMgr *ibclient.ObjectManager

type ipamCall struct {
	url string
	f   func(r interface{}) (map[string]interface{}, error)
	t   reflect.Type
}

func main() {
	defaultCidr := flag.String("default-cidr", "10.2.1.0/24", "Default Network CIDR if --subnet is not specified during docker network create")
	gridHostVar := flag.String("grid-host", "192.168.124.200", "IP of Infoblox Grid Host")
	wapiVerVar := flag.String("wapi-version", "2.0", "Infoblox WAPI Version.")
	wapiPortVar := flag.String("wapi-port", "443", "Infoblox WAPI Port.")
	globalNamespace := flag.String("global-view", "default", "Infoblox Network View for Global Address Space")
	localNamespace := flag.String("local-view", "default", "Infoblox Network View for Local Address Space")
	wapiUsername := flag.String("wapi-username", "", "Infoblox WAPI Username")
	wapiPassword := flag.String("wapi-password", "", "Infoblox WAPI Password")
	pluginDir := flag.String("plugin-dir", "/run/docker/plugins", "Docker plugin directory where driver socket is created")
	driverName := flag.String("driver-name", "mddi", "Name of Infoblox IPAM driver")

	flag.Parse()

	socketFile := setupSocket(*pluginDir, *driverName)
	log.Printf("Driver Name: '%s'", *driverName)
	log.Printf("Socket File: '%s'", socketFile)

	_, network, err := net.ParseCIDR(*defaultCidr)
	if err != nil {
		log.Panic(err)
	}
	np = network.String()
	log.Printf("Default Network CIDR: %s\n", np)

	conn := ibclient.NewConnector(
		*gridHostVar,
		*wapiVerVar,
		*wapiPortVar,
		*wapiUsername,
		*wapiPassword,
		false,
		"",
		120,
		100,
		100)

	dockerID, _ := getDockerID()
	if len(dockerID) > 0 {
		log.Printf("Docker id is '%s'\n", dockerID)
	}
	objMgr = ibclient.NewObjectManager(conn, *globalNamespace, *localNamespace, dockerID)

	ipamDrv := NewIpamDriver(objMgr, np)
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

			var req interface{}
			req = nil
			if c.t != nil {
				req = reflect.New(c.t).Interface()
				if err := json.NewDecoder(r.Body).Decode(req); err != nil {
					log.Printf("%s: Bad Request Error: %s\n", url, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			res, _ := c.f(req)

			log.Printf("res is '%s'\n", res)
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
