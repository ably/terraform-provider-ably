terraform {

  required_providers {
    ably = {
      source = "hashicorp/ably"
    }
  }
}

provider "ably" {
  token = <Control API token>
}
