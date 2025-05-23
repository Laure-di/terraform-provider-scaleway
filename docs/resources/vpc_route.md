---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_route"
---

# Resource: scaleway_vpc_route

Creates and manages Scaleway VPC Routes.
For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/).

## Example Usage

### With Instance

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-vpn"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "tf-pn-vpn"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_instance_server" "server01" {
  name  = "tf-server-vpn"
  type  = "PLAY2-MICRO"
  image = "openvpn"
}

resource "scaleway_instance_private_nic" "pnic01" {
  private_network_id = scaleway_vpc_private_network.pn01.id
  server_id          = scaleway_instance_server.server01.id
}

resource "scaleway_vpc_route" "rt01" {
  vpc_id              = scaleway_vpc.vpc01.id
  description         = "tf-route-vpn"
  tags                = ["tf", "route"]
  destination         = "10.0.0.0/24"
  nexthop_resource_id = scaleway_instance_private_nic.pnic01.id
}
```

### With Baremetal

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-vpn"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "tf-pn-vpn"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

data "scaleway_baremetal_os" "my_os" {
  zone    = "fr-par-2"
  name    = "Ubuntu"
  version = "22.04 LTS (Jammy Jellyfish)"
}

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-B112X-SSD"
}

data "scaleway_baremetal_option" "private_network" {
  zone = "fr-par-2"
  name = "Private Network"
}

data "scaleway_iam_ssh_key" "my_key" {
  name = "main"
}

resource "scaleway_baremetal_server" "my_server" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = data.scaleway_baremetal_os.my_os.os_id
  ssh_key_ids = [data.scaleway_iam_ssh_key.my_key.id]

  options {
    id = data.scaleway_baremetal_option.private_network.option_id
  }
  private_network {
    id = scaleway_vpc_private_network.pn01.id
  }
}

resource "scaleway_vpc_route" "rt01" {
  vpc_id              = scaleway_vpc.vpc01.id
  description         = "tf-route-vpn"
  tags                = ["tf", "route"]
  destination         = "10.0.0.0/24"
  nexthop_resource_id = scaleway_baremetal_server.my_server.private_network.0.mapping_id
}
```

## Argument Reference

The following arguments are supported:

- `vpc_id` - (Required) The VPC ID the route belongs to.
- `description` - (Optional) The route description.
- `tags` - (Optional) The tags to associate with the route.
- `destination` - (Optional) The destination of the route.
- `nexthop_resource_id` - (Optional) The ID of the nexthop resource.
- `nexthop_private_network_id` - (Optional) The ID of the nexthop private network.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the route.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the route is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the route.
- `created_at` - The date and time of the creation of the route (RFC 3339 format).
- `updated_at` - The date and time of the creation of the route (RFC 3339 format).

~> **Important:** routes' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

## Import

Routes can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_vpc_route.main fr-par/11111111-1111-1111-1111-111111111111
```
