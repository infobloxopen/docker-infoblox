package main

import (
	"errors"
	"fmt"
	ipamsapi "github.com/docker/libnetwork/ipams/remote/api"
	netlabel "github.com/docker/libnetwork/netlabel"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"log"
	"strconv"
	"strings"
)

type Container struct {
	NetworkContainer string // CIDR of Network Container
	ContainerObj     *ibclient.NetworkContainer
	exhausted        bool
}

type InfobloxAddressSpace struct {
	NetviewName  string // Network View Name
	PrefixLength uint   // Prefix Length
	Containers   []Container
}

type AddressSpaceScope int

const (
	GLOBAL AddressSpaceScope = iota
	LOCAL
)

type InfobloxDriver struct {
	objMgr              *ibclient.ObjectManager
	addressSpaceByScope map[AddressSpaceScope]*InfobloxAddressSpace
	addressSpaceByView  map[string]*InfobloxAddressSpace
}

func (ibDrv *InfobloxDriver) PluginActivate(r interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"Implements": []interface{}{
			"IpamDriver",
		}}, nil
}

func (ibDrv *InfobloxDriver) GetCapabilities(r interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"RequiresMACAddress": true}, nil
}

func (ibDrv *InfobloxDriver) GetDefaultAddressSpaces(r interface{}) (map[string]interface{}, error) {
	globalViewRef, localViewRef, err := ibDrv.objMgr.CreateDefaultNetviews(
		ibDrv.addressSpaceByScope[GLOBAL].NetviewName,
		ibDrv.addressSpaceByScope[LOCAL].NetviewName)

	return map[string]interface{}{"GlobalDefaultAddressSpace": globalViewRef, "LocalDefaultAddressSpace": localViewRef}, err
}

func getPrefixLength(cidr string) (prefixLength string) {
	parts := strings.Split(cidr, "/")
	return parts[1]
}

func (ibDrv *InfobloxDriver) RequestAddress(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.RequestAddressRequest)

	macAddr := v.Options[netlabel.MacAddress]
	if len(macAddr) == 0 {
		log.Printf("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.\n")
	}
	network := ibclient.BuildNetworkFromRef(v.PoolID)
	fixedAddr, _ := ibDrv.objMgr.AllocateIP(network.NetviewName, network.Cidr, macAddr)

	return map[string]interface{}{"Address": fmt.Sprintf("%s/%s", fixedAddr.IPAddress, getPrefixLength(network.Cidr))}, nil
}

func (ibDrv *InfobloxDriver) ReleaseAddress(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.ReleaseAddressRequest)
	log.Printf("Releasing Address '%s' from Pool '%s'\n", v.Address, v.PoolID)
	network := ibclient.BuildNetworkFromRef(v.PoolID)
	ref, _ := ibDrv.objMgr.ReleaseIP(network.NetviewName, v.Address)
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s'! *******\n", v.Address)
	}

	return map[string]interface{}{}, nil
}

func (ibDrv *InfobloxDriver) requestSpecificNetwork(netview string, pool string) (*ibclient.Network, error) {
	network, err := ibDrv.objMgr.GetNetwork(netview, pool)
	if network == nil {
		network, err = ibDrv.objMgr.CreateNetwork(netview, pool)
	}

	return network, err
}

func (ibDrv *InfobloxDriver) createNetworkContainer(netview string, pool string) (*ibclient.NetworkContainer, error) {
	container, err := ibDrv.objMgr.GetNetworkContainer(netview, pool)
	if container == nil {
		container, err = ibDrv.objMgr.CreateNetworkContainer(netview, pool)
	}

	return container, err
}

func nextAvailableContainer(addrSpace *InfobloxAddressSpace) *Container {
	for i, _ := range addrSpace.Containers {
		if !addrSpace.Containers[i].exhausted {
			return &addrSpace.Containers[i]
		}
	}

	return nil
}

func resetContainers(addrSpace *InfobloxAddressSpace) {
	for i, _ := range addrSpace.Containers {
		addrSpace.Containers[i].exhausted = false
	}
}

func (ibDrv *InfobloxDriver) allocateNetworkHelper(addrSpace *InfobloxAddressSpace, prefixLen uint) (network *ibclient.Network, err error) {
	if prefixLen == 0 {
		prefixLen = addrSpace.PrefixLength
	}
	container := nextAvailableContainer(addrSpace)
	for container != nil {
		log.Printf("Allocating network from Container:'%s'", container.NetworkContainer)
		if container.ContainerObj == nil {
			var err error
			container.ContainerObj, err = ibDrv.createNetworkContainer(addrSpace.NetviewName, container.NetworkContainer)
			if err != nil {
				return nil, err
			}
		}
		network, err = ibDrv.objMgr.AllocateNetwork(addrSpace.NetviewName, container.NetworkContainer, prefixLen)
		if network != nil {
			break
		}
		container.exhausted = true
		container = nextAvailableContainer(addrSpace)
	}

	return network, nil
}

func (ibDrv *InfobloxDriver) allocateNetwork(netview string, prefixLen uint) (network *ibclient.Network, err error) {
	addrSpace := ibDrv.addressSpaceByView[netview]

	network, err = ibDrv.allocateNetworkHelper(addrSpace, prefixLen)
	if network == nil {
		resetContainers(addrSpace)
		network, err = ibDrv.allocateNetworkHelper(addrSpace, prefixLen)
	}

	if network == nil {
		err = errors.New("Cannot allocate network in Address Space")
	}
	return
}

func (ibDrv *InfobloxDriver) RequestPool(r interface{}) (res map[string]interface{}, err error) {
	v := r.(*ipamsapi.RequestPoolRequest)
	log.Printf("RequestPoolRequest is '%v'\n", v)

	netviewName := ibclient.BuildNetworkViewFromRef(v.AddressSpace).Name

	var network *ibclient.Network
	if len(v.Pool) > 0 {
		network, err = ibDrv.requestSpecificNetwork(netviewName, v.Pool)
	} else {
		var prefixLen uint
		if opt, ok := v.Options["prefix-length"]; ok {
			if v, err := strconv.ParseUint(opt, 10, 8); err == nil {
				prefixLen = uint(v)
			}
		}
		network, err = ibDrv.allocateNetwork(netviewName, prefixLen)
	}

	if network != nil {
		res = map[string]interface{}{"PoolID": network.Ref, "Pool": network.Cidr}
	}
	return res, err
}

func (ibDrv *InfobloxDriver) ReleasePool(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.ReleasePoolRequest)

	if len(v.PoolID) > 0 {
		ref, _ := ibDrv.objMgr.DeleteLocalNetwork(v.PoolID, ibDrv.addressSpaceByScope[LOCAL].NetviewName)
		if len(ref) > 0 {
			log.Printf("Network %s deleted from Infoblox\n", v.PoolID)
		}
	}

	return map[string]interface{}{}, nil
}

func makeContainers(containerList string) []Container {
	containers := make([]Container, 0)

	parts := strings.Split(containerList, ",")
	for _, p := range parts {
		containers = append(containers, Container{p, nil, false})
	}

	return containers
}

func NewInfobloxDriver(objMgr *ibclient.ObjectManager,
	globalNetview string, globalNetworkContainer string, globalPrefixLength uint,
	localNetview string, localNetworkContainer string, localPrefixLength uint) *InfobloxDriver {

	globalContainers := makeContainers(globalNetworkContainer)
	localContainers := makeContainers(localNetworkContainer)

	globalAddressSpace := &InfobloxAddressSpace{
		globalNetview, globalPrefixLength, globalContainers}
	localAddressSpace := &InfobloxAddressSpace{
		localNetview, localPrefixLength, localContainers}

	return &InfobloxDriver{
		objMgr: objMgr,
		addressSpaceByScope: map[AddressSpaceScope]*InfobloxAddressSpace{
			GLOBAL: globalAddressSpace,
			LOCAL:  localAddressSpace,
		},
		addressSpaceByView: map[string]*InfobloxAddressSpace{
			globalNetview: globalAddressSpace,
			localNetview:  localAddressSpace,
		},
	}
}
