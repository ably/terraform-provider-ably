resource "ably_ingress_rule_postgres_outbox" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  target = {
    url                 = "postgres://${var.postgres_user}:${var.postgres_password}@${var.postgres_host}:5432/${var.postgres_db}"
    outbox_table_schema = "public"
    outbox_table_name   = "outbox"
    nodes_table_schema  = "public"
    nodes_table_name    = "nodes"
    ssl_mode            = "prefer"
    ssl_root_cert       = "-----BEGIN CERTIFICATE----- MIIFiTCCA3GgAwIBAgIUYO1Lomxzj7VRawWwEFiQht9OLpUwDQYJKoZIhvcNAQEL BQAwTDELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1pY2hpZ2FuMQ8wDQYDVQQHDAZX ...snip... TOfReTlUQzgpXRW5h3n2LVXbXQhPGcVitb88Cm2R8cxQwgB1VncM8yvmKhREo2tz 7Y+sUx6eIl4dlNl9kVrH1TD3EwwtGsjUNlFSZhg= -----END CERTIFICATE-----"
    primary_site        = "us-east-1-A"
  }
}
