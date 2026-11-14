terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.29"
    }
  }

  backend "s3" {
    key                         = "tachigiki.tfstate"
    region                      = "auto"
    use_path_style              = true
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
    skip_s3_checksum            = true
  }
}

provider "cloudflare" {}

variable "account_id" {
  type = string
}

variable "domain" {
  type = string
}

variable "github_client_id" {
  type = string
}

variable "github_client_secret" {
  type      = string
  sensitive = true
}

variable "allowed_emails" {
  type = list(string)
}

resource "cloudflare_r2_bucket" "tachigiki" {
  account_id = var.account_id
  name       = "tachigiki"
}

resource "cloudflare_zero_trust_access_identity_provider" "github" {
  account_id = var.account_id
  name       = "GitHub"
  type       = "github"
  config = {
    client_id     = var.github_client_id
    client_secret = var.github_client_secret
  }
}

resource "cloudflare_zero_trust_access_policy" "allowed" {
  account_id = var.account_id
  name       = "tachigiki allowed"
  decision   = "allow"
  include    = [for e in var.allowed_emails : { email = { email = e } }]
}

resource "cloudflare_zero_trust_access_application" "tachigiki" {
  account_id                = var.account_id
  name                      = "tachigiki"
  type                      = "self_hosted"
  domain                    = var.domain
  allowed_idps              = [cloudflare_zero_trust_access_identity_provider.github.id]
  auto_redirect_to_identity = true
  policies = [{
    id         = cloudflare_zero_trust_access_policy.allowed.id
    precedence = 1
  }]
}
