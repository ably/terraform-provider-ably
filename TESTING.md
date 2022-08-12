# Terraform Provider Testing Proposal

This document has been created for the purposes of discussing testing strategies for the Ably Terraform Provider.

It aims to provide a proposal of the types of testing we could potentially implement with our Terraform provider (https://github.com/ably/terraform-provider-ably) 

## Testing Types

### Terraform Acceptance Test Framework

Hashicorp already provides a good set of established processes & tools for testing custom providers. The Ably Terraform provider will make use of these in addition to Terratest, a testing tool created by Gruntwork.
Provider Acceptance Testing

As per https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests, Terraform already includes a framework for constructing acceptance tests. These tests perform CRUD operations on live resources in steps using the custom provider.
It should be noted that as live resources are provisioned, it is possible that dangling resources could be left behind after tests and we should account for this.

It is suggested that we make use of accounts on https://ably-dev.com/ and "https://staging-control.ably-dev.net/v1" Control API endpoint whilst setting up testing. 

It should be noted that we will need to ensure that Enterprise features are enabled so we can test all resources. This also relates to the testing of https://github.com/ably/ably-control-go 

### Unit Testing

We should also implement unit testing for small components of our provider code where network connections and full provider connections are not required.
The standard golang testing package will be used for these.

Conventions on provider testing can be found here - https://www.terraform.io/plugin/sdkv2/testing/unit-testing

Golang testing package - https://pkg.go.dev/testing 

### Terratest Testing

After implementing Unit & Acceptance testing, we should consider using Terratest for E2E testing. 
Terratest is a tool created by Gruntwork which primarily focuses on testing infrastructure tooling. It provides a neat way to conduct automated testing for Terraform. You could unit test resources/modules and/or conduct Integration / E2E tests for larger scale setups. 

Like the Terraform Acceptance Test framework, Terratest allows you to automate the entire provision/test/destroy & CRUD lifecycle stages that Terraform provides. 

There is also inbuilt API testing which we could use to independently validate responses from Ably’s Control API. 

We could also benefit from the inbuilt retry functionality, something which a tool like this definitely needs. 

A good guide to using Terratest for testing can be found here - https://www.youtube.com/watch?v=xhHOW0EF5u8 

https://terratest.gruntwork.io/docs/ 

### Testing Considerations

- Testing should be done in a sandbox environment
- Tests should allow the ability to run on workstations and CI.
