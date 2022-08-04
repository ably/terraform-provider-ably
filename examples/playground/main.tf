/**
* # Description
*
* This is a showcase of resources that can be provisioned using the Ably Terraform provider.
* Information on the Ably Control API can be found here - [Ably Control API Docs](https://ably.com/docs/api/control-api)
* Note that the ably provider source is "github.com/ably/ably", it's convention to drop the terraform-provider- prefix as
* provider source strings should not include it.
*/

terraform {

  required_providers {
    ably = {
      source = "github.com/ably/ably"
    }
  }
}

# You can provide your Ably Token & URL inline or use environment variables ABLY_ACCOUNT_TOKEN & ABLY_URL
provider "ably" {
  token = "INSERT"
  url   =  "https://control.ably.net/v1"
}
