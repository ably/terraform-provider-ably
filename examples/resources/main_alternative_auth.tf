terraform {

  required_providers {
    ably = {
      source = "ably/ably"
    }
  }
}

provider "ably" {
  token = "<Control API token>"
}
