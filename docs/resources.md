# Resources

## clusterbook_ip_assignment

Assigns an available IP from a network pool to a cluster. The provider automatically selects the first available IP.

### Example

```hcl
resource "clusterbook_ip_assignment" "ingress" {
  network_key = "10.31.105"
  cluster     = "mycluster"
  status      = "ASSIGNED"
  create_dns  = true
}

output "assigned_ip" {
  value = clusterbook_ip_assignment.ingress.ip_address
}
```

### Arguments

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `network_key` | string | yes | - | Network subnet prefix (e.g. `10.31.105`) |
| `cluster` | string | yes | - | Cluster name |
| `status` | string | no | `ASSIGNED` | Assignment status (`ASSIGNED` or `PENDING`) |
| `create_dns` | bool | no | `false` | Create PowerDNS A record |

### Computed Attributes

| Attribute | Description |
|-----------|-------------|
| `id` | Resource ID (`network_key/ip_digit`) |
| `ip_address` | The assigned IP address (e.g. `10.31.105.5`) |

### Lifecycle

All attributes force resource replacement — IP assignments are immutable. Changing any attribute will release the current IP and assign a new one.

- **Create**: Finds available IP, assigns to cluster
- **Read**: Verifies assignment via cluster info endpoint
- **Delete**: Releases the IP (and DNS record if `create_dns` was set)

---

## clusterbook_network

Manages a network IP pool in clusterbook.

### Example

```hcl
resource "clusterbook_network" "lab" {
  network_key = "10.31.106"
  ip_from     = 1
  ip_to       = 50
}
```

### Arguments

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `network_key` | string | yes | Network subnet prefix (e.g. `10.31.106`) |
| `ip_from` | int | yes | Start of IP range (last octet) |
| `ip_to` | int | yes | End of IP range (last octet) |

### Computed Attributes

| Attribute | Description |
|-----------|-------------|
| `id` | Same as `network_key` |
| `total` | Total IPs in pool |
| `available` | Available IPs |
| `assigned` | Assigned IPs |
| `pending` | Pending IPs |

### Lifecycle

All attributes force resource replacement — network pools are immutable.

- **Create**: Creates network with IP range
- **Read**: Refreshes pool statistics
- **Delete**: Removes the network pool
