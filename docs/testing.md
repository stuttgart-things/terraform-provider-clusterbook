# Testing

## Dev Build

Build the provider locally and configure a dev override so Terraform uses the local binary:

```bash
# Build
cd terraform-provider-clusterbook
go build -o terraform-provider-clusterbook .

# Create dev override config
cat > .terraformrc <<EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/stuttgart-things/clusterbook" = "/path/to/terraform-provider-clusterbook"
  }
  direct {}
}
EOF
```

## Test Config

Create a test directory with a Terraform config:

```hcl
terraform {
  required_providers {
    clusterbook = {
      source = "registry.terraform.io/stuttgart-things/clusterbook"
    }
  }
}

provider "clusterbook" {
  url = "http://clusterbook.movie-scripts2.sthings-vsphere.labul.sva.de"
}

# List all networks
data "clusterbook_networks" "all" {}

output "networks" {
  value = data.clusterbook_networks.all.networks
}

# Query cluster info
data "clusterbook_cluster" "skyami" {
  name = "skyami"
}

output "skyami_ips" {
  value = data.clusterbook_cluster.skyami.ips
}

# Assign an IP
resource "clusterbook_ip_assignment" "test" {
  network_key = "10.31.103"
  cluster     = "tf-test"
  status      = "ASSIGNED"
}

output "assigned_ip" {
  value = clusterbook_ip_assignment.test.ip_address
}
```

## Running

```bash
# Plan
TF_CLI_CONFIG_FILE=.terraformrc terraform plan

# Apply
TF_CLI_CONFIG_FILE=.terraformrc terraform apply -auto-approve

# Destroy
TF_CLI_CONFIG_FILE=.terraformrc terraform destroy -auto-approve
```

## Verified Test Results

Tested against live clusterbook at `movie-scripts2.sthings-vsphere.labul.sva.de`:

| Operation | Result |
|-----------|--------|
| `data.clusterbook_networks` | Listed networks `10.31.103` (8 IPs, 4 available) and `10.31.104` (5 IPs, 0 available) |
| `data.clusterbook_cluster` | Queried `skyami` — returned IP `10.31.103.5`, status `ASSIGNED` |
| `terraform apply` (ip_assignment) | Assigned `10.31.103.3` to cluster `tf-test` |
| `terraform destroy` | Released IP, cluster `tf-test` removed from clusterbook |

### Plan Output

```
data.clusterbook_networks.all: Read complete after 0s
data.clusterbook_cluster.skyami: Read complete after 0s [name=skyami]

Terraform will perform the following actions:

  # clusterbook_ip_assignment.test will be created
  + resource "clusterbook_ip_assignment" "test" {
      + cluster     = "tf-test"
      + create_dns  = false
      + id          = (known after apply)
      + ip_address  = (known after apply)
      + network_key = "10.31.103"
      + status      = "ASSIGNED"
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

### Apply Output

```
clusterbook_ip_assignment.test: Creating...
clusterbook_ip_assignment.test: Creation complete after 0s [id=10.31.103/3]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

assigned_ip = "10.31.103.3"
networks = tolist([
  {
    "assigned"    = 3
    "available"   = 4
    "network_key" = "10.31.103"
    "pending"     = 1
    "total"       = 8
  },
  {
    "assigned"    = 4
    "available"   = 0
    "network_key" = "10.31.104"
    "pending"     = 1
    "total"       = 5
  },
])
skyami_ips = tolist([
  {
    "ip"      = "10.31.103.5"
    "network" = "10.31.103"
    "status"  = "ASSIGNED"
  },
])
```

### Destroy Output

```
clusterbook_ip_assignment.test: Destroying... [id=10.31.103/3]
clusterbook_ip_assignment.test: Destruction complete after 0s

Destroy complete! Resources: 1 destroyed.
```

### API Verification

```bash
# After apply — IP assigned
$ curl -s clusterbook.../api/v1/clusters/tf-test | jq
{
  "cluster": "tf-test",
  "ips": [
    { "network": "10.31.103", "ip": "10.31.103.3", "status": "ASSIGNED" }
  ]
}

# After destroy — IP released
$ curl -s clusterbook.../api/v1/clusters/tf-test
{"error":"cluster not found"}
```
