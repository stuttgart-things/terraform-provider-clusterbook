data "clusterbook_cluster" "info" {
  name = "mycluster"
}

output "cluster_ips" {
  value = data.clusterbook_cluster.info.ips
}
