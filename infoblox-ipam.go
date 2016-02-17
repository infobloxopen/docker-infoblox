package main

import (
	"encoding/json"
	"flag"
	"fmt"
	apitypes "github.com/docker/engine-api/types"
	ipamsapi "github.com/docker/libnetwork/ipams/remote/api"
	netlabel "github.com/docker/libnetwork/netlabel"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
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

func fPluginActivate(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(&map[string]interface{}{
		"Implements": []interface{}{
			"IpamDriver",
		},
	}); err != nil {
		log.Printf("/Plugin.Activate Internal Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("/Plugin.Activate completed\n")
}

func fIpamDriverGetCapabilities(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(&map[string]interface{}{"RequiresMACAddress": true}); err != nil {
		log.Printf("/IpamDriver.GetCapabilities Internal Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("/IpamDriver.GetCapabilities completed\n")
}

func fIpamDriverGetDefaultAddressSpaces(w http.ResponseWriter, r *http.Request) {
	globalView, localView := objMgr.CreateDefaultNetviews()

	if err := json.NewEncoder(w).Encode(&map[string]interface{}{"LocalDefaultAddressSpace": localView.Name, "GlobalDefaultAddressSpace": globalView.Name}); err != nil {
		log.Printf("/IpamDriver.GetDefaultAddressSpaces Internal Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("/IpamDriver.GetDefaultAddressSpaces completed\n")
}

func fIpamDriverRequestPool(w http.ResponseWriter, r *http.Request) {
	var v ipamsapi.RequestPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		log.Printf("/IpamDriver.RequestPool Bad Request Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pool := np
	if len(v.Pool) > 0 {
		pool = v.Pool
	}
	netview := v.AddressSpace

	network, _ := objMgr.GetNetwork(netview, pool)
	if network == nil {
		network, _ = objMgr.CreateNetwork(netview, pool)
	}

	if err := json.NewEncoder(w).Encode(&map[string]interface{}{"PoolID": network.Ref, "Pool": network.Cidr}); err != nil {
		log.Printf("/IpamDriver.RequestPool Bad Response Error: %s\n", err)
	}

	log.Printf("/IpamDriver.RequestPool %#v completed\n", v)
}

func fIpamDriverRequestAddress(w http.ResponseWriter, r *http.Request) {
	var v ipamsapi.RequestAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		log.Printf("/IpamDriver.RequestAddress Bad Request Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	macAddr := v.Options[netlabel.MacAddress]
	if len(macAddr) == 0 {
		log.Printf("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.\n")
	}
	network := ibclient.BuildNetworkFromRef(v.PoolID)
	fixedAddr, _ := objMgr.AllocateIP(network.NetviewName, network.Cidr, macAddr)

	if err := json.NewEncoder(w).Encode(&map[string]interface{}{"Address": fmt.Sprintf("%s/24", fixedAddr.IPAddress)}); err != nil {
		log.Printf("/IpamDriver.RequestAddress Bad Response Error: %s\n", err)
	}
	log.Printf("/IpamDriver.RequestAddress %#v : %s completed\n", v, fixedAddr.IPAddress)
}

func fIpamDriverReleaseAddress(w http.ResponseWriter, r *http.Request) {
	var v ipamsapi.ReleaseAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		log.Printf("/IpamDriver.ReleaseAddress Bad Request Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("****** Release Address ******** PoolID: '%s', Address: '%s'", v.PoolID, v.Address)

	network := ibclient.BuildNetworkFromRef(v.PoolID)
	ref, _ := objMgr.ReleaseIP(network.NetviewName, v.Address)
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s'! *******\n", v.Address)
	}
	if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
		log.Printf("/IpamDriver.ReleaseAddress Bad Response Error: %s\n", err)
	}
	log.Printf("/IpamDriver.ReleaseAddress %s completed\n", v)
}

func fIpamDriverReleasePool(w http.ResponseWriter, r *http.Request) {
	var v ipamsapi.ReleasePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		log.Printf("/IpamDriver.ReleasePool Bad Request Error: %s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(v.PoolID) > 0 {
		ref, _ := objMgr.DeleteLocalNetwork(v.PoolID)
		if len(ref) > 0 {
			log.Printf("Network %s deleted from Infoblox\n", v.PoolID)
		}
	}
	if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
		log.Printf("/IpamDriver.ReleasePool Bad Response Error: %s\n", err)
	}
	log.Printf("/IpamDriver.ReleasePool %s completed\n", v.PoolID)
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

	var fp map[string]func(http.ResponseWriter, *http.Request) = make(map[string]func(http.ResponseWriter, *http.Request))
	fp["/Plugin.Activate"] = fPluginActivate
	fp["/IpamDriver.GetCapabilities"] = fIpamDriverGetCapabilities
	fp["/IpamDriver.GetDefaultAddressSpaces"] = fIpamDriverGetDefaultAddressSpaces
	fp["/IpamDriver.RequestPool"] = fIpamDriverRequestPool
	fp["/IpamDriver.RequestAddress"] = fIpamDriverRequestAddress
	fp["/IpamDriver.ReleaseAddress"] = fIpamDriverReleaseAddress
	fp["/IpamDriver.ReleasePool"] = fIpamDriverReleasePool

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Plugin: %s\n", r.URL.String())
		if h, ok := fp[r.URL.String()]; ok {
			h(w, r)
			return
		}
		fmt.Fprintf(w, "{ \"Error\": \"%s\"}", r.URL.String())

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
