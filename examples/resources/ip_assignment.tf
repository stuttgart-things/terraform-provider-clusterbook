resource "clusterbook_ip_assignment" "ingress" {
  network_key = "10.31.105"
  cluster     = "mycluster"
  status      = "ASSIGNED"
  create_dns  = true
}

output "assigned_ip" {
  value = clusterbook_ip_assignment.ingress.ip_address
}
