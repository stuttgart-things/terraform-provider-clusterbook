# Data Sources

## clusterbook_cluster

Query IP assignments for a specific cluster.

### Example

```hcl
data "clusterbook_cluster" "info" {
  name = "mycluster"
}

output "cluster_ips" {
  value = data.clusterbook_cluster.info.ips
}
```

### Arguments

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Cluster name |

### Computed Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `ips` | list | IP assignments for the cluster |
| `ips[].network` | string | Network prefix (e.g. `10.31.105`) |
| `ips[].ip` | string | Full IP address (e.g. `10.31.105.5`) |
| `ips[].status` | string | Assignment status |

---

## clusterbook_networks

List all network pools in clusterbook.

### Example

```hcl
data "clusterbook_networks" "all" {}

output "networks" {
  value = data.clusterbook_networks.all.networks
}
```

### Computed Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `networks` | list | All network pools |
| `networks[].network_key` | string | Network subnet prefix |
| `networks[].total` | float | Total IPs |
| `networks[].available` | float | Available IPs |
| `networks[].assigned` | float | Assigned IPs |
| `networks[].pending` | float | Pending IPs |
