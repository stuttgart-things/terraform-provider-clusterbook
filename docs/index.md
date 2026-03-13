# terraform-provider-clusterbook

Terraform provider for [clusterbook](https://github.com/stuttgart-things/clusterbook) IPAM. Manage IP address assignments and network pools as Terraform resources.

## Quick Start

```hcl
provider "clusterbook" {
  url = "https://clusterbook.example.com"
}

# Assign an IP to a cluster
resource "clusterbook_ip_assignment" "ingress" {
  network_key = "10.31.105"
  cluster     = "mycluster"
  create_dns  = true
}

output "ip" {
  value = clusterbook_ip_assignment.ingress.ip_address
}
```

## Provider Configuration

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | yes | Clusterbook API URL |

## Installation

Download the binary for your platform from [GitHub Releases](https://github.com/stuttgart-things/terraform-provider-clusterbook/releases) and place it in:

```
~/.terraform.d/plugins/registry.terraform.io/stuttgart-things/clusterbook/<VERSION>/<OS>_<ARCH>/
```

Add a `~/.terraformrc` to use the local plugin:

```hcl
provider_installation {
  filesystem_mirror {
    path = "~/.terraform.d/plugins"
  }
  direct {}
}
```
