package main

import (
	ibclient "github.com/infobloxopen/infoblox-go-client"
)

const (
	EA_GRID_CONFIG_GRID_SYNC_SUPPORT           = "Grid Sync Support"
	EA_GRID_CONFIG_GRID_SYNC_MINIMUM_WAIT_TIME = "Grid Sync Minimum Wait Time"
	EA_GRID_CONFIG_GRID_SYNC_MAXIMUM_WAIT_TIME = "Grid Sync Maximum Wait Time"
	EA_GRID_CONFIG_DEFAULT_NETWORK_VIEW_SCOPE  = "Default Network View Scope"
	EA_GRID_CONFIG_DEFAULT_NETWORK_VIEW        = "Default Network View"
	EA_GRID_CONFIG_DEFAULT_HOST_NAME_PATTERN   = "Default Host Name Pattern"
	EA_GRID_CONFIG_DEFAULT_DOMAIN_NAME_PATTERN = "Default Domain Name Pattern"
	EA_GRID_CONFIG_NS_GROUP                    = "NS Group"
	EA_GRID_CONFIG_DNS_VIEW                    = "DNS View"
	EA_GRID_CONFIG_NETWORK_TEMPLATE            = "Network Template"
	EA_GRID_CONFIG_ADMIN_NETWORK_DELETION      = "Admin Network Deletion"
	EA_GRID_CONFIG_IP_ALLOCATION_STRATEGY      = "IP Allocation Strategy"
	EA_GRID_CONFIG_DNS_RECORD_BINDING_TYPES    = "DNS Record Binding Types"
	EA_GRID_CONFIG_DNS_RECORD_UNBINDING_TYPES  = "DNS Record Unbinding Types"
	EA_GRID_CONFIG_DNS_RECORD_REMOVABLE_TYPES  = "DNS Record Removable Types"
	EA_GRID_CONFIG_DHCP_SUPPORT                = "DHCP Support"
	EA_GRID_CONFIG_DNS_SUPPORT                 = "DNS Support"
	EA_GRID_CONFIG_RELAY_SUPPORT               = "Relay Support"
	EA_GRID_CONFIG_USE_GM_FOR_DHCP             = "Use Grid Master for DHCP"
	EA_GRID_CONFIG_TENANT_NAME_PERSISTENCE     = "Tenant Name Persistence"
	EA_GRID_CONFIG_REPORT_GRID_SYNC_TIME       = "Report Grid Sync Time"
	EA_GRID_CONFIG_ALLOW_SERVICE_RESTART       = "Allow Service Restart"

	EA_LAST_GRID_SYNC_TIME = "Last Grid Sync Time"

	EA_MAPPING_ADDRESS_SCOPE_ID   = "Address Scope ID Mapping"
	EA_MAPPING_ADDRESS_SCOPE_NAME = "Address Scope Name Mapping"
	EA_MAPPING_TENANT_ID          = "Tenant ID Mapping"
	EA_MAPPING_TENANT_NAME        = "Tenant Name Mapping"
	EA_MAPPING_NETWORK_ID         = "Network ID Mapping"
	EA_MAPPING_NETWORK_NAME       = "Network Name Mapping"
	EA_MAPPING_SUBNET_ID          = "Subnet ID Mapping"
	EA_MAPPING_SUBNET_CIDR        = "Subnet CIDR Mapping"

	EA_CLOUD_ADAPTER_ID = "Cloud Adapter ID"
	EA_IS_CLOUD_MEMBER  = "Is Cloud Member"

	EA_SUBNET_ID             = "Subnet ID"
	EA_SUBNET_NAME           = "Subnet Name"
	EA_NETWORK_ID            = "Network ID"
	EA_NETWORK_NAME          = "Network Name"
	EA_NETWORK_ENCAP         = "Network Encap"
	EA_SEGMENTATION_ID       = "Segmentation ID"
	EA_PHYSICAL_NETWORK_NAME = "Physical Network Name"
	EA_PORT_ID               = "Port ID"
	EA_PORT_DEVICE_OWNER     = "Port Attached Device - Device Owner"
	EA_PORT_DEVICE_ID        = "Port Attached Device - Device ID"
	EA_VM_ID                 = "VM ID"
	EA_VM_NAME               = "VM Name"
	EA_IP_TYPE               = "IP Type"
	EA_TENANT_ID             = "Tenant ID"
	EA_TENANT_NAME           = "Tenant Name"
	EA_ACCOUNT               = "Account"
	EA_CLOUD_API_OWNED       = "Cloud API Owned"
	EA_CMP_TYPE              = "CMP Type"
	EA_IS_EXTERNAL           = "Is External"
	EA_IS_SHARED             = "Is Shared"
)

const (
	EA_TYPE_STRING  = "STRING"
	EA_TYPE_ENUM    = "ENUM"
	EA_TYPE_INTEGER = "INTEGER"
)

var RequiredEADefs = []ibclient.EADefinition{
	{Name: EA_CLOUD_API_OWNED, Type: EA_TYPE_ENUM, Flags: "C",
		ListValues: []ibclient.EADefListValue{"True", "False"},
		Comment:    "Is Cloud API owned"},
	{Name: EA_CMP_TYPE, Type: EA_TYPE_STRING, Flags: "C",
		Comment: "CMP Types (Docker)"},
	{Name: EA_TENANT_ID, Type: EA_TYPE_STRING, Flags: "C",
		Comment: "Docker Engine ID"},
	{Name: EA_VM_ID, Type: EA_TYPE_STRING, Flags: "C",
		Comment: "Containter ID in Docker"},
}

func GetRequiredEADefs() []ibclient.EADefinition {
	ea_defs := RequiredEADefs
	res := make([]ibclient.EADefinition, len(ea_defs))
	for i, d := range ea_defs {
		res[i] = *ibclient.NewEADefinition(d)
	}

	return res
}
