Configuration Guide
===================
This document describes how to configure vNIOS and the IPAM driver.

Introduction
------------
vNIOS is the Infoblox virtual appliance that you can download from the Infoblox Download Center:

- Point your browser to https://www.infoblox.com/infoblox-download-center
- Scroll down to the section "Network Service(DNS, DHCP, IPAM)"
- The product to download is "Infoblox DDI (DNS, DHCP, IPAM)". Click "Try it Now"
- This will take you through a brief registration screen.
- After registration is complete you will receive an email which includes a link that takes you to the "Product Evaluation Portal".    

Setting up vNIOS
----------------
Once you're in the "Product Evaluation Portal", you can find download links as well as instructional videos. It is
strongly recommended that you download the VMware version of the product as VMware is the platform on which the videos
are based.

- Under section "Required Downloads", download "Infoblox DDI" for VMware.
- After download is complete, scroll down to section "Setup and Installation Videos"
- Follow the video instruction: "Video 1: Infoblox Cloud Network Automation Installation and Setup"
- Follow the video to completion, as the instruction to activate vNIOS "Cloud Network Automation" feature is in the later part of the video. (You can however skip over section on configuring DHCP and DNS, as well as section on "vRealization Orchestrator".

The following additional steps are required:
- You need to give cloud-api admin user permission to create and modify DNS Views. Instructions on how to add permission to "cloud-api-only" group is included in the video. Follow the same instructions to add "All DNS Views" permission under the "DNS Permissions" Permssion Type.

Configuring Cloud Extensible Attributes using create-ea-defs tool
-----------------------------------------------------------------
If the "Cloud Network Automation" license is activated, then the Cloud Extensible Attributes used by the docker-ipam-plugin
can be defined using the create-ea-defs tool in the infoblox/docker-ipam-tools docker image.

To run create-ea-defs:
```
docker run infoblox/container-ipam-tool:0.0.1 create-ea-defs --debug --grid-host 10.120.21.150 --wapi-username=admin --wapi-password=infoblox --wapi-version=2.3
```

To use the configuration file for create-ea-defs:
```
docker run -v /etc/infoblox:/etc/infoblox infoblox/container-ipam-tool:0.0.1 create-ea-defs --debug --conf-file docker-infoblox.conf
```

IPAM Driver Configuration
-------------------------
Based on the vNIOS configuration, update the following driver configuration:
- Set grid-host to the management IP address of vNIOS
- Set username and password to that for the Cloud Admin user on vNIOS.

These configurations can be applied by editing the "run.sh" and "run-container.sh" shell scripts.
