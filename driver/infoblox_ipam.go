package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	ipamApi "github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/libnetwork/netlabel"
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

func (ibDrv *InfobloxDriver) GetCapabilities() (*ipamApi.CapabilitiesResponse, error) {
	logrus.Infof("GetCapabilities called")
	return &ipamApi.CapabilitiesResponse{RequiresMACAddress: true}, nil
}

func (ibDrv *InfobloxDriver) GetDefaultAddressSpaces() (*ipamApi.AddressSpacesResponse, error) {
	logrus.Infof("GetDefaultAddressSpaces called")
	globalViewRef, localViewRef, err := ibDrv.objMgr.CreateDefaultNetviews(
		ibDrv.addressSpaceByScope[GLOBAL].NetviewName,
		ibDrv.addressSpaceByScope[LOCAL].NetviewName)

	return &ipamApi.AddressSpacesResponse{LocalDefaultAddressSpace: localViewRef, GlobalDefaultAddressSpace: globalViewRef}, err
}

func getPrefixLength(cidr string) (prefixLength string) {
	parts := strings.Split(cidr, "/")
	return parts[1]
}

func (ibDrv *InfobloxDriver) RequestAddress(r *ipamApi.RequestAddressRequest) (*ipamApi.RequestAddressResponse, error) {
	network := ibclient.BuildNetworkFromRef(r.PoolID)

	macAddr := r.Options[netlabel.MacAddress]
	if len(macAddr) == 0 {
		macAddr = ibclient.MACADDR_ZERO
		log.Printf("RequestAddressRequest contains empty MAC Address. '%s' will be used.\n", macAddr)
	}

	fixedAddr, _ := ibDrv.objMgr.GetFixedAddress(network.NetviewName, network.Cidr, "", macAddr)
	if fixedAddr != nil {
		if r.Address != "" {
			if fixedAddr.IPAddress != r.Address {
				msg := fmt.Sprintf("Requested MAC '%s' is already associated with a difference IP '%s' (requested: '%s')",
					macAddr, fixedAddr.IPAddress, r.Address)
				log.Printf("RequestAddress: %s", msg)
				return &ipamApi.RequestAddressResponse{}, errors.New(msg)
			}
		}
	}

	var err error
	if fixedAddr == nil {
		fixedAddr, err = ibDrv.objMgr.AllocateIP(network.NetviewName, network.Cidr, r.Address, macAddr, "")
	}

	var res ipamApi.RequestAddressResponse
	if fixedAddr == nil || err != nil {
		res = ipamApi.RequestAddressResponse{}
	} else {
		res = ipamApi.RequestAddressResponse{Address: fmt.Sprintf("%s/%s", fixedAddr.IPAddress, getPrefixLength(network.Cidr))}
	}

	return &res, err
}

func (ibDrv *InfobloxDriver) ReleaseAddress(r *ipamApi.ReleaseAddressRequest) error {
	log.Printf("Releasing Address '%s' from Pool '%s'\n", r.Address, r.PoolID)
	network := ibclient.BuildNetworkFromRef(r.PoolID)
	ref, _ := ibDrv.objMgr.ReleaseIP(network.NetviewName, network.Cidr, r.Address, "")
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s'! *******\n", r.Address)
	}

	return nil
}

func (ibDrv *InfobloxDriver) requestSpecificNetwork(netview string, pool string, networkName string) (*ibclient.Network, error) {
	network, err := ibDrv.objMgr.GetNetwork(netview, pool, nil)
	if err != nil {
		return nil, err
	}
	if network != nil {
		if n, ok := network.Ea["Network Name"]; !ok || n != networkName {
			msg := fmt.Sprintf("Network (%s) already in use", network.Cidr)
			log.Printf("requestSpecificNetwork: %s", msg)
			return nil, errors.New(msg)
		}
	} else {
		networkByName, err := ibDrv.objMgr.GetNetwork(netview, "", ibclient.EA{"Network Name": networkName})
		if err != nil {
			return nil, err
		}
		if networkByName != nil {
			if networkByName.Cidr != pool {
				msg := fmt.Sprintf("Network name (%s) has different CIDR (%s)", networkName, networkByName.Cidr)
				log.Printf("requestSpecificNetwork: %s", msg)
				return nil, errors.New(msg)
			}
		}
	}

	if network == nil {
		network, err = ibDrv.objMgr.CreateNetwork(netview, pool, networkName)
		log.Printf("requestSpecificNetwork: CreateNetwork returns '%s', err='%s'", *network, err)
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

func (ibDrv *InfobloxDriver) allocateNetworkHelper(addrSpace *InfobloxAddressSpace, prefixLen uint, networkName string) (network *ibclient.Network, err error) {
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
		network, err = ibDrv.objMgr.AllocateNetwork(addrSpace.NetviewName, container.NetworkContainer, prefixLen, networkName)
		if network != nil {
			break
		}
		container.exhausted = true
		container = nextAvailableContainer(addrSpace)
	}

	return network, nil
}

func (ibDrv *InfobloxDriver) allocateNetwork(netview string, prefixLen uint, networkName string) (network *ibclient.Network, err error) {
	addrSpace := ibDrv.addressSpaceByView[netview]

	network, err = ibDrv.allocateNetworkHelper(addrSpace, prefixLen, networkName)
	if network == nil {
		resetContainers(addrSpace)
		network, err = ibDrv.allocateNetworkHelper(addrSpace, prefixLen, networkName)
	}

	if network == nil {
		err = errors.New("Cannot allocate network in Address Space")
	}
	return
}

func (ibDrv *InfobloxDriver) RequestPool(r *ipamApi.RequestPoolRequest) (res *ipamApi.RequestPoolResponse, err error) {
	log.Printf("RequestPoolRequest is '%v'\n", r)

	netviewName := ibclient.BuildNetworkViewFromRef(r.AddressSpace).Name

	var network *ibclient.Network
	var networkName string

	if opt, ok := r.Options["network-name"]; ok {
		networkName = opt
	}

	if len(r.Pool) > 0 {
		network, err = ibDrv.requestSpecificNetwork(netviewName, r.Pool, networkName)
	} else {
		var prefixLen uint
		var networkByName *ibclient.Network
		if networkName != "" {
			networkByName, err = ibDrv.objMgr.GetNetwork(netviewName, "", ibclient.EA{"Network Name": networkName})
			if err != nil {
				return
			}
		}
		if networkByName != nil {
			log.Printf("RequestNetwork: GetNetwork by name returns '%s'", *networkByName)
			network = networkByName
		} else {
			if opt, ok := r.Options["prefix-length"]; ok {
				if v, err := strconv.ParseUint(opt, 10, 8); err == nil {
					prefixLen = uint(v)
				}
			}
			network, err = ibDrv.allocateNetwork(netviewName, prefixLen, networkName)
		}
	}

	if network != nil {
		res = &ipamApi.RequestPoolResponse{PoolID: network.Ref, Pool: network.Cidr}
	}
	return
}

func (ibDrv *InfobloxDriver) ReleasePool(r *ipamApi.ReleasePoolRequest) error {

	if len(r.PoolID) > 0 {
		networkFromRef := ibclient.BuildNetworkFromRef(r.PoolID)
		network, err := ibDrv.objMgr.GetNetwork(networkFromRef.NetviewName, networkFromRef.Cidr, nil)
		if err != nil {
			return err
		}

		// if network has a valid looking "Network Name" EA, assume that
		// it is shared with others - hence not deleted.
		if n, ok := network.Ea["Network Name"]; ok && n != "" {
			return nil
		}

		ref, _ := ibDrv.objMgr.DeleteNetwork(r.PoolID, networkFromRef.NetviewName)
		if len(ref) > 0 {
			log.Printf("Network %s successfully deleted from Infoblox\n", r.PoolID)
		}
	}

	return nil
}

func makeContainers(containerList string) []Container {
	var containers []Container

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
