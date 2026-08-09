terraform {
  required_version = ">= 1.11"

  required_providers {
    mazevault = {
      source  = "MazeVault/mazevault"
      version = "~> 1.0"
    }
  }
}

# Authentication via API token
provider "mazevault" {
  server_url = "https://vault.example.com"
  api_token  = var.mazevault_token
}

# Alternatively, use OAuth2 client credentials:
# provider "mazevault" {
#   server_url    = "https://vault.example.com"
#   client_id     = var.client_id
#   client_secret = var.client_secret
# }

variable "mazevault_token" {
  type      = string
  sensitive = true
}

variable "org_id" {
  type = string
}
