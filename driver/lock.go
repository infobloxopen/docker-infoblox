package main

import (
	"github.com/Sirupsen/logrus"
	ibclient "github.com/infobloxopen/infoblox-go-client"
	"time"
)

type GridLocker struct {
	timeout time.Time
	objMgr  *ibclient.ObjectManager
	name    string
	ref     string
}

func (l *GridLocker) Lock() bool {
	// create EA definition with name as in lockName variable
	// If EA is created successfully, then lock is acquired
	// If response has an error, then lock is held by some other resource

	logrus.Debugf("Creating lock %s\n", l.name)

	eaDef := ibclient.EADefinition{Name: l.name, Type: "STRING", Flags: "C",
		Comment:      "Docker Resource Lock",
		DefaultValue: string(time.Now().Unix())}

	// TODO: implement timeout so that lock is never held infinitely for a resource
	// This can happen when a system which acquired lock crashes without releasing the lock

	eaDefRef, err := l.objMgr.CreateEADefinition(eaDef)

	if err != nil {
		// Already locked
		logrus.Debugf("Failed to create lock. %s\n", err)
		return false
	}
	l.ref = eaDefRef.Ref
	return true
}

func (l *GridLocker) UnLock() error {
	_, err := l.objMgr.DeleteEADefinition(l.ref)
	return err
}
