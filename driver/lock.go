package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/infobloxopen/docker-infoblox/common"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"time"
)

const (
	timeout int32 = 120 // in seconds
)

type NVLocker struct {
	name   string
	objMgr *ibclient.ObjectManager
}

func (l *NVLocker) createLockRequest() []*ibclient.RequestBody {

	req := []*ibclient.RequestBody{
		&ibclient.RequestBody{
			Method: "GET",
			Object: "networkview",
			Data: map[string]interface{}{
				"name": l.name,
				"*" + common.EA_DOCKER_PLUGIN_LOCK: "Available",
			},
			Args: map[string]string{
				"_return_fields": "extattrs",
			},
			AssignState: map[string]string{
				"NET_VIEW_REF": "_ref",
			},
			Discard: true,
		},
		&ibclient.RequestBody{
			Method: "PUT",
			Object: "##STATE:NET_VIEW_REF:##",
			Data: map[string]interface{}{
				"extattrs+": map[string]interface{}{
					common.EA_DOCKER_PLUGIN_LOCK: map[string]string{
						"value": l.objMgr.TenantID,
					},
					common.EA_DOCKER_PLUGIN_LOCK_TIME: map[string]int32{
						"value": int32(time.Now().Unix()),
					},
				},
			},
			EnableSubstitution: true,
			Discard:            true,
		},
		&ibclient.RequestBody{
			Method: "GET",
			Object: "##STATE:NET_VIEW_REF:##",
			Args: map[string]string{
				"_return_fields": "extattrs",
			},
			AssignState: map[string]string{
				"DOCKER-ID": "*" + common.EA_DOCKER_PLUGIN_LOCK,
			},
			EnableSubstitution: true,
			Discard:            true,
		},
		&ibclient.RequestBody{
			Method: "STATE:DISPLAY",
		},
	}

	return req
}

func (l *NVLocker) createUnlockRequest(force bool) []*ibclient.RequestBody {

	getData := map[string]interface{}{"name": l.name}
	if !force {
		getData["*"+common.EA_DOCKER_PLUGIN_LOCK] = l.objMgr.TenantID
	}

	req := []*ibclient.RequestBody{
		&ibclient.RequestBody{
			Method: "GET",
			Object: "networkview",
			Data:   getData,
			Args: map[string]string{
				"_return_fields": "extattrs",
			},
			AssignState: map[string]string{
				"NET_VIEW_REF": "_ref",
			},
			Discard: true,
		},
		&ibclient.RequestBody{
			Method: "PUT",
			Object: "##STATE:NET_VIEW_REF:##",
			Data: map[string]interface{}{
				"extattrs+": map[string]interface{}{
					common.EA_DOCKER_PLUGIN_LOCK: map[string]string{
						"value": "Available",
					},
				},
			},
			EnableSubstitution: true,
			Discard:            true,
		},
		&ibclient.RequestBody{
			Method: "PUT",
			Object: "##STATE:NET_VIEW_REF:##",
			Data: map[string]interface{}{
				"extattrs-": map[string]interface{}{
					common.EA_DOCKER_PLUGIN_LOCK_TIME: map[string]interface{}{},
				},
			},
			EnableSubstitution: true,
			Discard:            true,
		},
		&ibclient.RequestBody{
			Method: "GET",
			Object: "##STATE:NET_VIEW_REF:##",
			Args: map[string]string{
				"_return_fields": "extattrs",
			},
			AssignState: map[string]string{
				"DOCKER-ID": "*" + common.EA_DOCKER_PLUGIN_LOCK,
			},
			EnableSubstitution: true,
			Discard:            true,
		},
		&ibclient.RequestBody{
			Method: "STATE:DISPLAY",
		},
	}

	return req
}

func (l *NVLocker) Lock() bool {

	logrus.Debugf("Creating lock on network niew%s\n", l.name)

	req := l.createLockRequest()
	res, err := l.objMgr.CreateMultiObjectRequest(req)

	if err != nil {
		logrus.Debugf("Failed to create lock on network view %s: %s", l.name, err)

		//Check for timeout
		nw, err := l.objMgr.GetNetworkView(l.name)
		if err != nil {
			logrus.Debugf("Failed to get the network view object for %s : %s", l.name, err)
			return false
		}

		if t, ok := nw.Ea[common.EA_DOCKER_PLUGIN_LOCK_TIME]; ok {
			if int32(time.Now().Unix())-int32(t.(int)) > timeout {
				logrus.Debugf("Lock is timed out. Forcefully acquiring it.")
				//remove the lock forcefully and acquire it
				l.UnLock(true)
				// try to get lock again
				return l.Lock()
			}
		}
		return false
	}

	dockerID := res[0]["DOCKER-ID"]
	if dockerID == l.objMgr.TenantID {
		logrus.Debugln("Got the lock !!!")
		return true
	}

	return false
}

func (l *NVLocker) UnLock(force bool) bool {
	// To unlock set the Docker-Plugin-Lock EA of network view to Available and
	// remove the Docker-Plugin-Lock-Time EA

	req := l.createUnlockRequest(force)
	res, err := l.objMgr.CreateMultiObjectRequest(req)

	if err != nil {
		logrus.Debugf("Failed to release lock on network view %s: %s", l.name, err)
		return false
	}

	dockerID := res[0]["DOCKER-ID"]
	if dockerID == "Available" {
		logrus.Debugln("Removed the lock!")
		return true
	}

	return false
}
