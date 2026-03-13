data "clusterbook_networks" "all" {}

output "all_networks" {
  value = data.clusterbook_networks.all.networks
}
