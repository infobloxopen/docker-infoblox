package main

import (
	"fmt"
	ipamsapi "github.com/docker/libnetwork/ipams/remote/api"
	netlabel "github.com/docker/libnetwork/netlabel"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"log"
)

type IpamDriver struct {
	objMgr      *ibclient.ObjectManager
	defaultPool string
}

func (ipamDrv *IpamDriver) PluginActivate(r interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"Implements": []interface{}{
			"IpamDriver",
		}}, nil
}

func (ipamDrv *IpamDriver) GetCapabilities(r interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"RequiresMACAddress": true}, nil
}

func (ipamDrv *IpamDriver) GetDefaultAddressSpaces(r interface{}) (map[string]interface{}, error) {
	globalView, localView := ipamDrv.objMgr.CreateDefaultNetviews()

	return map[string]interface{}{"LocalDefaultAddressSpace": localView.Name, "GlobalDefaultAddressSpace": globalView.Name}, nil
}

func (ipamDrv *IpamDriver) RequestAddress(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.RequestAddressRequest)

	macAddr := v.Options[netlabel.MacAddress]
	if len(macAddr) == 0 {
		log.Printf("RequestAddressRequest contains empty MAC Address. '00:00:00:00:00:00' will be used.\n")
	}
	network := ibclient.BuildNetworkFromRef(v.PoolID)
	fixedAddr, _ := ipamDrv.objMgr.AllocateIP(network.NetviewName, network.Cidr, macAddr)

	return map[string]interface{}{"Address": fmt.Sprintf("%s/24", fixedAddr.IPAddress)}, nil
}

func (ipamDrv *IpamDriver) ReleaseAddress(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.ReleaseAddressRequest)

	network := ibclient.BuildNetworkFromRef(v.PoolID)
	ref, _ := ipamDrv.objMgr.ReleaseIP(network.NetviewName, v.Address)
	if ref == "" {
		log.Printf("***** IP Cannot be deleted '%s'! *******\n", v.Address)
	}

	return map[string]interface{}{}, nil
}

func (ipamDrv *IpamDriver) RequestPool(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.RequestPoolRequest)

	pool := ipamDrv.defaultPool
	if len(v.Pool) > 0 {
		pool = v.Pool
	}
	netview := v.AddressSpace

	network, _ := ipamDrv.objMgr.GetNetwork(netview, pool)
	if network == nil {
		network, _ = ipamDrv.objMgr.CreateNetwork(netview, pool)
	}

	return map[string]interface{}{"PoolID": network.Ref, "Pool": network.Cidr}, nil
}

func (ipamDrv *IpamDriver) ReleasePool(r interface{}) (map[string]interface{}, error) {
	v := r.(*ipamsapi.ReleasePoolRequest)

	if len(v.PoolID) > 0 {
		ref, _ := ipamDrv.objMgr.DeleteLocalNetwork(v.PoolID)
		if len(ref) > 0 {
			log.Printf("Network %s deleted from Infoblox\n", v.PoolID)
		}
	}

	return map[string]interface{}{}, nil
}

func NewIpamDriver(objMgr *ibclient.ObjectManager, defaultPool string) *IpamDriver {
	return &IpamDriver{objMgr: objMgr, defaultPool: defaultPool}
}
