package ably_control

import (
	"fmt"
	"testing"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var cert string = `-----BEGIN CERTIFICATE-----
MIIGGDCCBQCgAwIBAgIIN67hazCzGwwwDQYJKoZIhvcNAQELBQAwgZYxCzAJBgNV
BAYTAlVTMRMwEQYDVQQKDApBcHBsZSBJbmMuMSwwKgYDVQQLDCNBcHBsZSBXb3Js
ZHdpZGUgRGV2ZWxvcGVyIFJlbGF0aW9uczFEMEIGA1UEAww7QXBwbGUgV29ybGR3
aWRlIERldmVsb3BlciBSZWxhdGlvbnMgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkw
HhcNMTcxMjE0MTMyMTQzWhcNMTkwMTEzMTMyMTQzWjCBkzEhMB8GCgmSJomT8ixk
AQEMEWlvLmFibHkucHVzaC1kZW1vMS8wLQYDVQQDDCZBcHBsZSBQdXNoIFNlcnZp
Y2VzOiBpby5hYmx5LnB1c2gtZGVtbzETMBEGA1UECwwKWFhZOThBVkRSNjEbMBkG
A1UECgwSQWJseSBSZWFsLXRpbWUgTHRkMQswCQYDVQQGEwJVUzCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBAN1xIAgCXDRzRiw2lzQjRPS/RNQrLceq+qMR
cU2quAmeBL+nR68TCtwcnEYzyE7kiDRd/sFxV6gH3raCYf7YjyGVVaLmG4WfK3gh
qWeuaNap29lerkwTcRhBTdJVzvCbffO/4QuzbNkpuFv91p99FFVtlsPGBTR4THRy
iLv1fOeA+p8+wbOEeyQah7OS70AwJ4p1UPQF57rd6eR8ea/tbgvB4LAYo4qSAbSP
Uizz2boYofAlhbgVg+eJbcEDCy1Z6UOT4bv9T8Uet7GUnDcLA5W9YhAZu5gS9hrZ
HhK/p4n+QLVkvgigVgWIEAaxqniIaFslOOR1LFH9yioYZFFDM38CAwEAAaOCAmkw
ggJlMB0GA1UdDgQWBBRiwxd3NoXz1hczNuRDAzJ15Jk/ADAMBgNVHRMBAf8EAjAA
MB8GA1UdIwQYMBaAFIgnFwmpthhgi+zruvZHWcVSVKO3MIIBHAYDVR0gBIIBEzCC
AQ8wggELBgkqhkiG92NkBQEwgf0wgcMGCCsGAQUFBwICMIG2DIGzUmVsaWFuY2Ug
b24gdGhpcyBjZXJ0aWZpY2F0ZSBieSBhbnkgcGFydHkgYXNzdW1lcyBhY2NlcHRh
bmNlIG9mIHRoZSB0aGVuIGFwcGxpY2FibGUgc3RhbmRhcmQgdGVybXMgYW5kIGNv
bmRpdGlvbnMgb2YgdXNlLCBjZXJ0aWZpY2F0ZSBwb2xpY3kgYW5kIGNlcnRpZmlj
YXRpb24gcHJhY3RpY2Ugc3RhdGVtZW50cy4wNQYIKwYBBQUHAgEWKWh0dHA6Ly93
d3cuYXBwbGUuY29tL2NlcnRpZmljYXRlYXV0aG9yaXR5MDAGA1UdHwQpMCcwJaAj
oCGGH2h0dHA6Ly9jcmwuYXBwbGUuY29tL3d3ZHJjYS5jcmwwDgYDVR0PAQH/BAQD
AgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMCMBAGCiqGSIb3Y2QGAwEEAgUAMBAGCiqG
SIb3Y2QGAwIEAgUAMHoGCiqGSIb3Y2QGAwYEbDBqDBFpby5hYmx5LnB1c2gtZGVt
bzAFDANhcHAMFmlvLmFibHkucHVzaC1kZW1vLnZvaXAwBgwEdm9pcAweaW8uYWJs
eS5wdXNoLWRlbW8uY29tcGxpY2F0aW9uMA4MDGNvbXBsaWNhdGlvbjANBgkqhkiG
9w0BAQsFAAOCAQEAsir24YYCpb4/Pn76ur2kn38I1Nfe4gsxqNfkmItOYKc0H0UX
JOGBaoJTrvKGb5jaMXwsDTuUlgO1tz1tOKZk31Y1YUlmqnZvw8YOJp1GgEB/8kJF
9BNvsoJOOy0AFuW7IV67OfPCBzA9j8JoNHz+6hiHILycF0LEvR7Cs8qE0rFFJLqW
T+LHFRs+0vqDDxxriaEXbJGeRNAseUrT2O4ey41U2GaBwyuo41SFfF0jgS2fINku
OYA5r5WJTBbKHaxALnanHqDDWYRn/RXjtdDjdOby3JvFdyZul75lqBnKQ+2n6F7g
umtfr8jT+b4KanXzXe/TJ7OTrxH2nwgvKgnj9w==
-----END CERTIFICATE-----`

var key string = "-----BEGIN PRIVATE KEY-----\naaa\n-----END PRIVATE KEY-----"

// Test Create and Update of an Ably app with:
// Step 1: Create w/ params (name=autogenerated, status=enabled, tls_only=true)
// Step 2: Update w/ params (name=acc-test-{autogenerated}, status=disabled, tls_only=false)
func TestAccAblyApp(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	update_app_name := "acc-test-" + app_name
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyAppConfig(&ably_control_go.App{
					Name:                   app_name,
					Status:                 "enabled",
					TLSOnly:                true,
					FcmKey:                 "a",
					ApnsCertificate:        cert,
					ApnsPrivateKey:         key,
					ApnsUseSandboxEndpoint: true,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_app.app0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_app.app0", "tls_only", "true"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyAppConfig(&ably_control_go.App{
					Name:                   update_app_name,
					Status:                 "disabled",
					TLSOnly:                false,
					FcmKey:                 "b",
					ApnsCertificate:        cert,
					ApnsPrivateKey:         key,
					ApnsUseSandboxEndpoint: true,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_app.app0", "status", "disabled"),
					resource.TestCheckResourceAttr("ably_app.app0", "tls_only", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Create App with status = disabled. This should fail and return status = enabled - Issue known and fix being worked on
// For now, the test will be commented out
// TODO: Verify fix with this test and update Doc Comment
// func TestAccAblyAppDisabledStatus(t *testing.T) {
// 	app_name := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:  func() { testAccPreCheck(t) },
// 		Providers: testAccProviders,
// 		Steps: []resource.TestStep{
// 			// Create and Read testing of ably_app.app0
// 			{
// 				Config: testAccAblyAppConfig(app_name, "disabled", "false"),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
// 					resource.TestCheckResourceAttr("ably_app.app0", "status", "disabled"),
// 					resource.TestCheckResourceAttr("ably_app.app0", "tls_only", "false"),
// 				),
// 			},
// 		},
// 	})
// }

// Function with inline HCL to provision an ably_app resource
// Takes App name, status and tls_only status as function params.
func testAccAblyAppConfig(app *ably_control_go.App) string {
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
	name                      = %[1]q
	status                    = %[2]q
	tls_only                  = %[3]t
	fcm_key                   = %[4]q
	apns_certificate          = %[5]q
	apns_private_key          = %[6]q
	apns_use_sandbox_endpoint = %[7]t


}
`, app.Name, app.Status, app.TLSOnly, app.FcmKey, app.ApnsCertificate, app.ApnsPrivateKey, app.ApnsUseSandboxEndpoint)
}
