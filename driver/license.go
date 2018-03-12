package main

import (
	"fmt"
	"strings"

	ibclient "github.com/infobloxopen/infoblox-go-client"
	"github.com/sirupsen/logrus"
)

//Checks for cloud license in nios
func CheckForCloudLicense(objMgr *ibclient.ObjectManager) {
	err := CheckLicense(objMgr, "cloud")
	if err != nil {
		logrus.Fatal("Check Cloud License : ", err)
	}
}

func CheckLicense(objMgr *ibclient.ObjectManager, licenseType string) (err error) {
	license, err := objMgr.GetLicense()
	if err != nil {
		return
	}
	for _, v := range license {
		if strings.ToLower(v.Licensetype) == licenseType {
			if v.ExpirationStatus != "DELETED" && v.ExpirationStatus != "EXPIRED" {
				return
			}
		}
	}
	err = fmt.Errorf("%s License not available/applied. Apply the license and try again", licenseType)
	return
}
