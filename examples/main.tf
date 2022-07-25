/**
* # Description
*
* This is a showcase of resources that can be provisioned using the Ably Terraform provider.
* Information on the Ably Control API can be found here - [Ably Control API Docs](https://ably.com/docs/api/control-api)
*/

terraform {

  required_providers {
    ably = {
      source = "hashicorp/ably"
    }
  }
}
