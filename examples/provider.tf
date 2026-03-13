terraform {
  required_providers {
    clusterbook = {
      source = "stuttgart-things/clusterbook"
    }
  }
}

provider "clusterbook" {
  url = "http://localhost:8080"
}
