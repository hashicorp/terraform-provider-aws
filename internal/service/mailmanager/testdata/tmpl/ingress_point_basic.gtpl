resource "aws_mailmanager_ingress_point" "test" {
{{- template "region" }}
  name              = var.rName
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id
{{- template "tags" . }}
}

resource "aws_mailmanager_traffic_policy" "test" {
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
}

resource "aws_mailmanager_rule_set" "test" {
  name = var.rName

  rule {
    action {
      add_header {
        header_name  = "X-Test"
        header_value = "example"
      }
    }
  }
}
