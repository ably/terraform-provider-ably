terraform {

  required_providers {
    ably = {
      source = "github.com/ably/ably"
    }
  }
}

# You can provide your Ably Token & URL inline or use environment variables ABLY_ACCOUNT_TOKEN & ABLY_URL
# provider "ably" {
#   token = <Control API token>
# }
