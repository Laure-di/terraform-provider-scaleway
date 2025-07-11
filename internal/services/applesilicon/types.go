package applesilicon

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandPrivateNetworks(pn any) map[string]*[]string {
	privateNetworks := make(map[string]*[]string)

	for _, op := range pn.(*schema.Set).List() {
		rawPN := op.(map[string]any)
		id := locality.ExpandID(rawPN["id"].(string))

		ipamIPIDs := &[]string{}

		if ipamIPs, ok := rawPN["ipam_ip_ids"]; ok && ipamIPs != nil {
			ipamIPsList := ipamIPs.([]any)
			if len(ipamIPsList) > 0 {
				ips := make([]string, len(ipamIPsList))

				for i, ip := range ipamIPsList {
					ips[i] = locality.ExpandID(ip.(string))
				}

				ipamIPIDs = &ips
			}
		}

		privateNetworks[id] = ipamIPIDs
	}

	return privateNetworks
}

func flattenPrivateNetworks(region scw.Region, privateNetworks []*applesilicon.ServerPrivateNetwork) any {
	flattenedPrivateNetworks := []map[string]any(nil)
	for _, privateNetwork := range privateNetworks {
		flattenedPrivateNetworks = append(flattenedPrivateNetworks, map[string]any{
			"id":          regional.NewIDString(region, privateNetwork.PrivateNetworkID),
			"ipam_ip_ids": regional.NewIDStrings(region, privateNetwork.IpamIPIDs),
			"vlan":        types.FlattenUint32Ptr(privateNetwork.Vlan),
			"status":      privateNetwork.Status,
			"created_at":  types.FlattenTime(privateNetwork.CreatedAt),
			"updated_at":  types.FlattenTime(privateNetwork.UpdatedAt),
		})
	}

	return flattenedPrivateNetworks
}
