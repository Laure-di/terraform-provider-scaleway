---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_public_gateway_pat_rule"
---

# Resource: scaleway_vpc_public_gateway_pat_rule

Creates and manages Scaleway Public Gateway PAT (Port Address Translation).
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/public-gateway/#pat-rules-e75d10).

## Example Usage

```terraform
resource "scaleway_instance_security_group" "sg01" {
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"

  inbound_rule {
    action   = "accept"
    port     = 22
    protocol = "TCP"
  }
}

resource "scaleway_instance_server" "srv01" {
  name              = "my-server"
  type              = "PLAY2-NANO"
  image             = "ubuntu_jammy"
  security_group_id = scaleway_instance_security_group.sg01.id
}

resource "scaleway_instance_private_nic" "pnic01" {
  server_id          = scaleway_instance_server.srv01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "my-pn"
}

resource "scaleway_vpc_public_gateway_dhcp" "dhcp01" {
  subnet = "192.168.0.0/24"
}

resource "scaleway_vpc_public_gateway_ip" "ip01" {}

resource "scaleway_vpc_public_gateway" "pg01" {
  name  = "my-pg"
  type  = "VPC-GW-S"
  ip_id = scaleway_vpc_public_gateway_ip.ip01.id
}

resource "scaleway_vpc_gateway_network" "gn01" {
  gateway_id         = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  dhcp_id            = scaleway_vpc_public_gateway_dhcp.dhcp01.id
  cleanup_dhcp       = true
  enable_masquerade  = true
}

resource "scaleway_vpc_public_gateway_dhcp_reservation" "rsv01" {
  gateway_network_id = scaleway_vpc_gateway_network.gn01.id
  mac_address        = scaleway_instance_private_nic.pnic01.mac_address
  ip_address         = "192.168.0.7"
}

# PAT rule for SSH traffic
resource "scaleway_vpc_public_gateway_pat_rule" "pat01" {
  gateway_id   = scaleway_vpc_public_gateway.pg01.id
  private_ip   = scaleway_vpc_public_gateway_dhcp_reservation.rsv01.ip_address
  private_port = 22
  public_port  = 2202
  protocol     = "tcp"
}
```

## Argument Reference

The following arguments are supported:

- `gateway_id` - (Required) The ID of the Public Gateway.
- `private_ip` - (Required) The private IP address to forward data to.
- `public_port` - (Required) The public port to listen on.
- `private_port` - (Required) The private port to translate to.
- `protocol` - (Defaults to both) The protocol the rule should apply to. Possible values are `both`, `tcp` and `udp`.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Public Gateway DHCP configuration should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Public Gateway PAT rule.

~> **Important:** Public Gateway PAT rule IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `organization_id` - The Organization ID the PAT rule configuration is associated with.
- `created_at` - The date and time of the creation of the PAT rule configuration.
- `updated_at` - The date and time of the last update of the PAT rule configuration.

## Import

Public Gateway PAT rule configurations can be imported using `{zone}/{id}`, e.g.

```bash
terraform import scaleway_vpc_public_gateway_pat_rule.main fr-par-1/11111111-1111-1111-1111-111111111111
```
