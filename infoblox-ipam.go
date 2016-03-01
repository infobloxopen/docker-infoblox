package main

import (
	"fmt"
	ipamsapi "github.com/docker/libnetwork/ipams/remote/api"
	netlabel "github.com/docker/libnetwork/netlabel"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"log"
)

type InfobloxDriver struct {
	objMgr      *ibclient.ObjectManager
	globalNetview string
	localNetview string
	defaultPool string
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
	globalViewRef, localViewRef := ibDrv.objMgr.CreateDefaultNetviews(ibDrv.globalNetview, ibDrv.localNetview)

	return map[string]interface{}{"GlobalDefaultAddressSpace": globalViewRef, "LocalDefaultAddressSpace": localViewRef}, nil
}

func (ibDrv *InfobloxDriver) RequestAddress(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.RequestAddressRequest)

	macAddr := v.Options[netlabel.MacAddress]
	if len(macAddr) == 0 {
		log.Printf("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.\n")
	}
	network := ibclient.BuildNetworkFromRef(v.PoolID)
	fixedAddr, _ := ibDrv.objMgr.AllocateIP(network.NetviewName, network.Cidr, macAddr)

	return map[string]interface{}{"Address": fmt.Sprintf("%s/24", fixedAddr.IPAddress)}, nil
}

func (ibDrv *InfobloxDriver) ReleaseAddress(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.ReleaseAddressRequest)

	network := ibclient.BuildNetworkFromRef(v.PoolID)
	ref, _ := ibDrv.objMgr.ReleaseIP(network.NetviewName, v.Address)
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s'! *******\n", v.Address)
	}

	return map[string]interface{}{}, nil
}

func (ibDrv *InfobloxDriver) RequestPool(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.RequestPoolRequest)

	pool := ibDrv.defaultPool
	if len(v.Pool) > 0 {
		pool = v.Pool
	}
	netview := ibclient.BuildNetworkViewFromRef(v.AddressSpace).Name

	network, _ := ibDrv.objMgr.GetNetwork(netview, pool)
	if network == nil {
		network, _ = ibDrv.objMgr.CreateNetwork(netview, pool)
	}

	return map[string]interface{}{"PoolID": network.Ref, "Pool": network.Cidr}, nil
}

func (ibDrv *InfobloxDriver) ReleasePool(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.ReleasePoolRequest)

	if len(v.PoolID) > 0 {
		ref, _ := ibDrv.objMgr.DeleteLocalNetwork(v.PoolID, ibDrv.localNetview)
		if len(ref) > 0 {
			log.Printf("Network %s deleted from Infoblox\n", v.PoolID)
		}
	}

	return map[string]interface{}{}, nil
}

func NewInfobloxDriver(objMgr *ibclient.ObjectManager, globalNetview string, 
	localNetview string, defaultPool string) *InfobloxDriver {
	return &InfobloxDriver{objMgr: objMgr, globalNetview: globalNetview,
		localNetview: localNetview, defaultPool: defaultPool}
}
