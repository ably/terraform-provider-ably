resource "ably_ingress_rule_mongodb" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  target = {
    url        = "mongodb://${var.mongodb_user}:${var.mongodb_password}@${var.mongodb_host}:27017"
    database   = "coconut"
    collection = "coconut"
    pipeline = jsonencode([
      {
        "$set" = {
          "_ablyChannel" = "myChannel"
        }
      }
    ])
    full_document               = "off"
    full_document_before_change = "off"
    primary_site                = "us-east-1-A"
  }
}
