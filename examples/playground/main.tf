terraform {

  required_providers {
    ably = {
      source  = "ably/ably"
      version = ">=0.2.0"
    }
  }
}

# You can provide your Ably Token & URL inline or use environment variables ABLY_ACCOUNT_TOKEN & ABLY_URL
# provider "ably" {
#   token = <Control API token>
# }
