package main

import (
	"strings"

	ibclient "github.com/infobloxopen/infoblox-go-client"
	"github.com/sirupsen/logrus"
)

//Checks for cloud license in nios
func CheckForCloudLicense(objMgr *ibclient.ObjectManager) {
	flag, err := CheckLicense(objMgr, "cloud")
	if err != nil {
		logrus.Fatal("error", err)
	}
	if !flag {
		logrus.Fatal("Cloud License not available in Infoblox Appliance. Update and try again..")
	}
}

func CheckLicense(objMgr *ibclient.ObjectManager, licenseType string) (flag bool, err error) {
	license, err := objMgr.GetLicense()
	if err != nil {
		return flag, err
	}
	for _, v := range license {
		if strings.ToLower(v.Licensetype) == licenseType {
			if v.ExpirationStatus != "DELETED" && v.ExpirationStatus != "EXPIRED" {
				flag = true
			}
		}
	}
	return flag, err
}
