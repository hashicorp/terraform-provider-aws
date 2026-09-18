resource "aws_mailmanager_traffic_policy" "test" {
{{- template "region" }}
  default_action = "ALLOW"
  name           = var.rName

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
{{- template "tags" . }}
}
