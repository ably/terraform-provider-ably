package ably_control

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRulePulsar(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	update_app_name := "acc-test-" + app_name

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRulePulsarConfig(
					app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					"test-key",
					"my-tenant/my-namespace/my-topic",
					"pulsar://pulsar.us-west.example.com:6650/",
					"true",
					"json",
					"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.routing_key", "test-key"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.topic", "my-tenant/my-namespace/my-topic"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.service_url", "pulsar://pulsar.us-west.example.com:6650/"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.format", "json"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.authentication.token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRulePulsarConfig(
					update_app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					"test-key1",
					"my-tenant/my-namespace/my-topic1",
					"pulsar://pulsar.us-east.example.com:6650/",
					"false",
					"msgpack",
					"YWxnOkhTNTEyIHR5cDpKV1QK",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.routing_key", "test-key1"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.topic", "my-tenant/my-namespace/my-topic1"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.service_url", "pulsar://pulsar.us-east.example.com:6650/"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.format", "msgpack"),
					resource.TestCheckResourceAttr("ably_rule_pulsar.rule0", "target.authentication.token", "YWxnOkhTNTEyIHR5cDpKV1QK"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRulePulsarConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetRoutingKey string,
	targetTopic string,
	targetServiceURL string,
	targetEnveloped string,
	targetFormat string,
	targetAuthToken string,
) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
		source = "github.com/ably/ably"
		}
	}
}
	
# You can provide your Ably Token & URL inline or use environment variables ABLY_ACCOUNT_TOKEN & ABLY_URL
provider "ably" {}
	  
resource "ably_app" "app0" {
	name     = %[1]q
	status   = "enabled"
	tls_only = true
}

resource "ably_rule_pulsar" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}
	request_mode = %[5]q
	target = {
	  routing_key =  %[6]q
	  topic = %[7]q
	  service_url = %[8]q
	  tls_trust_certs = ["-----BEGIN CERTIFICATE-----\nMIIFqDCCA5ACCQDLT5J/mNSX4TANBgkqhkiG9w0BAQsFADCBlTELMAkGA1UEBhMC\nR0IxDzANBgNVBAgMBkxvbmRvbjEPMA0GA1UEBwwGTG9uZG9uMRAwDgYDVQQKDAdU\nZXN0T3JnMQwwCgYDVQQLDANTREsxIzAhBgNVBAMMGnB1bHNhci51cy13ZXN0LmV4\nYW1wbGUuY29tMR8wHQYJKoZIhvcNAQkBFhB0ZXN0QGV4YW1wbGUuY29tMB4XDTIy\nMDkyNjEwNTU0MFoXDTIzMDkyNjEwNTU0MFowgZUxCzAJBgNVBAYTAkdCMQ8wDQYD\nVQQIDAZMb25kb24xDzANBgNVBAcMBkxvbmRvbjEQMA4GA1UECgwHVGVzdE9yZzEM\nMAoGA1UECwwDU0RLMSMwIQYDVQQDDBpwdWxzYXIudXMtd2VzdC5leGFtcGxlLmNv\nbTEfMB0GCSqGSIb3DQEJARYQdGVzdEBleGFtcGxlLmNvbTCCAiIwDQYJKoZIhvcN\nAQEBBQADggIPADCCAgoCggIBAMx7TXLIvBh+CQDat9PIUlTLAFSR6KAJe5j659O1\n17Lyue2QcnpOTYAf5QOYyvNmC91l/KoAlPVr6DRig2vZAB/cHDans95+CRJfzA8r\npTwHbT2C9a14tTY0T4E5GAEGEBU7tv5fgfD0smwTtv6eiJ4In9EzQO3p0OLOsAeD\npxTnLoLSTMoTUTgg3v5A8BBtzTb3lI3+HDxGe8anb5c5cVirRca5KSkQNZR+QBPg\n9KF6RTsEKhdq+ptteHFbIEw0cM5MitEyeWFmG2kf4V3SX+8+Ntrf1EopGenRCJEj\nbZit8vOPI43kgP0mGHOzoQQRnhGyTNmjtE+Z2xxEzs7eYSXQa8kxO+kb37mAwRuX\nRhAfsL8oj8Hxs+UmJk1F+XJIma37F/JBW671R+L7vZmJE3OLM53IwmtrELFoxLsi\noc8urBM6onSe5ZxZ8B+VLGkVZpTJ1PWeqbKCsp6RCweuOxZXb5M0kz6HewXwLKNK\nt4A4CqIfEngZR74HuH0r9G1Ql4YkhkWCsj4+9b5Uq21d/aZU2C3wTbbPWJ9khqVT\nNjWWi78FyoC1HCjWYgKCK1SQsYcqhq2nWj+MbqMN13k5Qc85hjFcmCB1SiWH+gv5\nXQLUbXZAN4DKuN/iVGLM33teBPp7yVZpZbNfdaTAQWLWiw1ROUEcyKRt+B2rp+F+\nxP5VAgMBAAEwDQYJKoZIhvcNAQELBQADggIBADraBsNjnURUF6Zn/gTpF2nlqzhf\nBYOhlUv+6k9q5IJlqYtFT7mo+EhWf8xbso4vWipEJPy85DTG7gr/P1gJC4FBIaOe\nR0WIlwZukz/S1W9KJ4eeh3b92QjYn+Sbx1Mc8qUaZk45MsLZrpSyHsrbvXGQsDwj\nCyRAexJN7gGMBteHMgfZGQINQe3Lya76rl2xPM4jsd8mWWASwT715fSiqRbWi7a/\nXbTP/ENtUj5PRrHliXFL+6nCQa6y73Qdt2o3Ob6ZWlFywv3of2wKas1bYdE1ZxKw\n5Br9/m1hhxrH52AnDuR9BfNIp3Z/eCFCXLI0WHsxBEEgPZfUmo5iwRKWrPVcVwLS\nlTNCPTuMG/Fnl+MbXtvu30bVjLrH54yKFQEv39cPP9OpuC/YW/nW56eR3h4MLmqP\njX4y7IOBkUAczjZHPsfMM8DcemUYcswIjTtk8piz9YPDo3qNQGsnZNba4uDulQ0U\nrfEDa9HzWB6hiJ02g+XssiSo9mbann0qU0ZWmCxiBDN5eMQYJ//RMZym0ccAu9Ug\nxapS7YtDmqkq2FQdj++IFst0ktBvXDV8AVz4MuZwY9adSZFmwHonHomiLfgySPRR\ncYzK74pwWRa5PWLzBXHU9oC316izLXBQO4OhUdJtaqwqNd22L4UFinQcJL12Up5c\n4XIgNCFSXyfq8ZGj\n-----END CERTIFICATE-----\n"]
	  enveloped = %[9]s
	  format = %[10]q
	  authentication = {
		mode = "token"
		token = %[11]q
	  }
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetRoutingKey, targetTopic, targetServiceURL, targetEnveloped, targetFormat, targetAuthToken)
}
