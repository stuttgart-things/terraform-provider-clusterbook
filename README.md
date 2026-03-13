# terraform-provider-clusterbook

Terraform provider for [clusterbook](https://github.com/stuttgart-things/clusterbook) IPAM.

## Provider Configuration

```hcl
provider "clusterbook" {
  url = "https://clusterbook.example.com"
}
```

## Resources

### clusterbook_ip_assignment

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

### clusterbook_network

Creates a network IP pool.

```hcl
resource "clusterbook_network" "lab" {
  network_key = "10.31.106"
  ip_from     = 1
  ip_to       = 50
}
```

## Data Sources

### clusterbook_cluster

Query IP assignments for a cluster.

```hcl
data "clusterbook_cluster" "info" {
  name = "mycluster"
}
```

### clusterbook_networks

List all network pools.

```hcl
data "clusterbook_networks" "all" {}
```

## Development

```bash
go build -o terraform-provider-clusterbook .
```
