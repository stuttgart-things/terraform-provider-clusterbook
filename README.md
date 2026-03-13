# stuttgart-things/terraform-provider-clusterbook

Terraform provider for clusterbook IPAM

<div align="center">
  <p>
    <img src="https://github.com/stuttgart-things/docs/blob/main/hugo/sthings-argo.png" alt="sthings" width="450" />
  </p>
  <p>
    Terraform provider for <a href="https://github.com/stuttgart-things/clusterbook">clusterbook</a> IPAM
  </p>
  <p>
    <a href="https://github.com/stuttgart-things/terraform-provider-clusterbook/releases"><img src="https://img.shields.io/github/v/release/stuttgart-things/terraform-provider-clusterbook?style=for-the-badge" alt="Release"></a>
    <a href="https://stuttgart-things.github.io/terraform-provider-clusterbook"><img src="https://img.shields.io/badge/docs-pages-blue?style=for-the-badge" alt="Pages"></a>
    <a href="https://github.com/stuttgart-things/terraform-provider-clusterbook/releases"><img src="https://img.shields.io/github/downloads/stuttgart-things/terraform-provider-clusterbook/total?style=for-the-badge" alt="Downloads"></a>
  </p>
</div>

## PROVIDER CONFIGURATION

```hcl
provider "clusterbook" {
  url = "https://clusterbook.example.com"
}
```

## RESOURCES

<details><summary>clusterbook_ip_assignment</summary>

Assigns an available IP from a network pool to a cluster.

```hcl
resource "clusterbook_ip_assignment" "ingress" {
  network_key = "10.31.105"
  cluster     = "mycluster"
  status      = "ASSIGNED"    # optional, default: ASSIGNED
  create_dns  = true          # optional, default: false
}

output "assigned_ip" {
  value = clusterbook_ip_assignment.ingress.ip_address
}
```

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `network_key` | string | yes | - | Network subnet prefix |
| `cluster` | string | yes | - | Cluster name |
| `status` | string | no | `ASSIGNED` | `ASSIGNED` or `PENDING` |
| `create_dns` | bool | no | `false` | Create PowerDNS A record |

**Computed**: `ip_address`, `id`

</details>

<details><summary>clusterbook_network</summary>

Creates a network IP pool.

```hcl
resource "clusterbook_network" "lab" {
  network_key = "10.31.106"
  ip_from     = 1
  ip_to       = 50
}
```

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `network_key` | string | yes | Network subnet prefix |
| `ip_from` | int | yes | Start of IP range (last octet) |
| `ip_to` | int | yes | End of IP range (last octet) |

**Computed**: `total`, `available`, `assigned`, `pending`

</details>

## DATA SOURCES

<details><summary>clusterbook_cluster</summary>

Query IP assignments for a cluster.

```hcl
data "clusterbook_cluster" "info" {
  name = "mycluster"
}
```

</details>

<details><summary>clusterbook_networks</summary>

List all network pools.

```hcl
data "clusterbook_networks" "all" {}
```

</details>

## INSTALLATION

Download the binary for your platform from [Releases](https://github.com/stuttgart-things/terraform-provider-clusterbook/releases) and place it in:

```
~/.terraform.d/plugins/registry.terraform.io/stuttgart-things/clusterbook/<VERSION>/<OS>_<ARCH>/
```

## DEVELOPMENT

```bash
go build -o terraform-provider-clusterbook .
```

## LICENSE

Licensed under the Apache License 2.0.
