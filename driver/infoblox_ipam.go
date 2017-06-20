package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	ipamApi "github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/libnetwork/netlabel"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"math/rand"
	"strconv"
	"strings"
	"time"
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
	logrus.Infof("GetCapabilities called\n")
	return &ipamApi.CapabilitiesResponse{RequiresMACAddress: true}, nil
}

func (ibDrv *InfobloxDriver) GetDefaultAddressSpaces() (*ipamApi.AddressSpacesResponse, error) {
	logrus.Infof("GetDefaultAddressSpaces called\n")
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
	logrus.Debugf("RequestAddress Called : %+v\n", r)

	network := ibclient.BuildNetworkFromRef(r.PoolID)
	macAddr := r.Options[netlabel.MacAddress]
	var fixedAddr *ibclient.FixedAddress
	var err error
	// TODO: Do not create multiple gateways for shared networks

	// Gateway request does not have a MAC address in it. If RequestAddressType is Gateway
	// then allocate IP with MAC 00:00:00:00:00:00
	// If RequestAddressType is not Gateway and MAC is empty then its overlay network
	// request for the Gateway.
	if rType, ok := r.Options["RequestAddressType"]; macAddr == "" || (ok && rType == netlabel.Gateway) {
		logrus.Debugln("Request for Gateway IP")

		if len(macAddr) == 0 {
			macAddr = ibclient.MACADDR_ZERO
			logrus.Infof("RequestAddressRequest contains empty MAC Address. '%s' will be used.\n", macAddr)
		}
		fixedAddr, err = ibDrv.objMgr.AllocateIP(network.NetviewName, network.Cidr, r.Address, macAddr, "")
		if err != nil {
			msg := fmt.Sprintf("Failed to allocate Gateway IP for network %s : %s", network.Cidr, err)
			return &ipamApi.RequestAddressResponse{}, errors.New(msg)
		}
	} else {
		// This request is for the container
		fixedAddr, err = ibDrv.objMgr.AllocateIP(network.NetviewName, network.Cidr, r.Address, macAddr, "")
		if err != nil {
			msg := fmt.Sprintf("Failed to allocate IP from pool %s for container with MAC %s : %s", network.Cidr, macAddr, err)
			return &ipamApi.RequestAddressResponse{}, errors.New(msg)
		}
	}
	logrus.Debugf("Allocated IP %s\n", fixedAddr.IPAddress)
	res := ipamApi.RequestAddressResponse{Address: fmt.Sprintf("%s/%s", fixedAddr.IPAddress, getPrefixLength(network.Cidr))}
	return &res, nil
}

func (ibDrv *InfobloxDriver) ReleaseAddress(r *ipamApi.ReleaseAddressRequest) error {
	logrus.Debugf("ReleaseAddress Called : %+v\n", r)
	logrus.Infof("Releasing Address '%s' from Pool '%s'\n", r.Address, r.PoolID)
	network := ibclient.BuildNetworkFromRef(r.PoolID)
	ref, _ := ibDrv.objMgr.ReleaseIP(network.NetviewName, network.Cidr, r.Address, "")
	if ref == "" {
		logrus.Warnf("IP Cannot be deleted '%s'!\n", r.Address)
	}

	return nil
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
		logrus.Infof("Allocating network from Container:'%s'\n", container.NetworkContainer)
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

func (ibDrv *InfobloxDriver) allocateNetwork(netView string, prefixLen uint, networkName string) (network *ibclient.Network, err error) {
	addrSpace := ibDrv.addressSpaceByView[netView]

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

func (ibDrv *InfobloxDriver) getSharedNetwork(netViewName string, pool string, networkName string) (*ibclient.Network, error) {

	// Check if this network exists in NIOS. If it exists then check if pool is also
	// passed. If so then CIDR should match for both the networks.
	logrus.Infof("Searching network with name (%s) in view (%s)\n", networkName, netViewName)
	networkByName, err := ibDrv.objMgr.GetNetwork(netViewName, "", ibclient.EA{"Network Name": networkName})
	if err != nil {
		msg := fmt.Sprintf("Error while fetching shared network with name %s from NIOS: (%s)", networkName, err)
		return nil, errors.New(msg)
	}

	// If the name matches and request also specifies a pool
	// then verify if the networks have same CIDR as in the request Pool.
	if networkByName != nil {
		if len(pool) > 0 {
			if networkByName.Cidr != pool {
				msg := fmt.Sprintf("Cannot allocate Pool %s. Network name %s is already allocated to Network %s", pool, networkName, networkByName.Cidr)
				logrus.Errorln(msg)
				return nil, errors.New(msg)
			}
		}
		// Network exists and is same as requested
		logrus.Debugf("Shared network (%s) found with CIDR (%s)\n", networkName, networkByName.Cidr)
		return networkByName, nil
	}
	return networkByName, nil
}

func (ibDrv *InfobloxDriver) createSharedNetwork(netViewName string, pool string, networkName string, prefixLen uint) (*ibclient.Network, error) {

	// Network View + Network Name will always be unique for a shared network
	lockName := netViewName + networkName
	l := GridLocker{name: lockName, objMgr: ibDrv.objMgr}
	retryCount := 0
	for {
		// Get lock to create the network
		lock := l.Lock()
		if lock == false {
			// resource is held by other plugin instance. Wait for some time and
			// retry it again
			if retryCount >= 10 {
				return nil, fmt.Errorf("Failed to create network %s. Cannot get the lock.", networkName)
			}

			retryCount++
			logrus.Debugf("Failed to get the lock. Retry count %d out of 10.\n", retryCount)
			// sleep for random time (between 1 - 5 seconds) to reduce collisions
			time.Sleep(time.Duration(rand.Intn(4)+1) * time.Second)
			continue
		} else {
			// Got the lock. Create the network.
			logrus.Debugf("Got the lock %s. Creating the network %s\n", lockName, networkName)
			defer l.UnLock()
			break
		}
	}

	// get the network if it exists
	networkByName, err := ibDrv.getSharedNetwork(netViewName, pool, networkName)

	if err != nil {
		return nil, err
	}

	if networkByName != nil {
		return networkByName, nil
	}

	logrus.Infof("Shared Network %s not found. Creating shared network with CIDR %s\n", networkName, pool)
	network, err := ibDrv.createNetwork(netViewName, pool, networkName, prefixLen)
	return network, err
}

func (ibDrv *InfobloxDriver) createNetwork(netViewName string, pool string, networkName string, prefixLen uint) (network *ibclient.Network, err error) {

	if len(pool) > 0 {
		// Create the specific network sent in the RequestPool
		network, err = ibDrv.objMgr.CreateNetwork(netViewName, pool, networkName)
	} else {
		// Allocate from the network container
		network, err = ibDrv.allocateNetwork(netViewName, prefixLen, networkName)
	}
	return
}

func (ibDrv *InfobloxDriver) RequestPool(r *ipamApi.RequestPoolRequest) (*ipamApi.RequestPoolResponse, error) {
	logrus.Infof("RequestPool Called : %+v\n", r)

	netViewName := ibclient.BuildNetworkViewFromRef(r.AddressSpace).Name

	var networkName string
	var network *ibclient.Network
	var err error

	if opt, ok := r.Options["network-name"]; ok {
		networkName = opt
	}

	var prefixLen uint
	if opt, ok := r.Options["prefix-length"]; ok {
		if v, err := strconv.ParseUint(opt, 10, 8); err == nil {
			prefixLen = uint(v)
		}
	}

	if networkName != "" {
		// Create the shared network
		network, err = ibDrv.createSharedNetwork(netViewName, r.Pool, networkName, prefixLen)
	} else {
		// Check if this network exists. If network exists then throw error because
		// this is not a shared network.
		if len(r.Pool) > 0 {
			network, err = ibDrv.objMgr.GetNetwork(netViewName, r.Pool, nil)
			if err != nil {
				msg := fmt.Sprintf("Error while fetching network with pool %s from NIOS: (%s)", r.Pool, err)
				return nil, errors.New(msg)
			}
			if network != nil {
				msg := fmt.Sprintf("Network %s already exists in Network view %s", network.Cidr, netViewName)
				logrus.Errorln(msg)
				return nil, errors.New(msg)
			}
		}

		network, err = ibDrv.createNetwork(netViewName, r.Pool, "", prefixLen)
	}

	if err != nil {
		msg := fmt.Sprintf("Error while allocating pool: (%s)", err)
		return nil, errors.New(msg)
	} else {
		logrus.Infof("Network Allocated : %s \n", network.Cidr)
		return &ipamApi.RequestPoolResponse{PoolID: network.Ref, Pool: network.Cidr}, nil
	}
}

func (ibDrv *InfobloxDriver) ReleasePool(r *ipamApi.ReleasePoolRequest) error {
	logrus.Infof("ReleasePool Called : %+v\n", r)

	logrus.Debugf("Releasing Network %s\n", r.PoolID)
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
			logrus.Infof("Network %s successfully deleted from Infoblox\n", r.PoolID)
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
