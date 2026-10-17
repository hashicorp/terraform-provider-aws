resource "aws_networksecuritymanager_deployment" "test" {
{{- template "region" }}
  name        = var.rName
  policy_arns = [aws_networksecuritymanager_policy.test.arn]
  scope_arn   = aws_networksecuritymanager_scope.test.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
{{- template "tags" . }}
}

resource "aws_networksecuritymanager_policy" "test" {
{{- template "region" }}
  name          = var.rName
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.test.arn
  }
}

resource "aws_networksecuritymanager_rule" "test" {
{{- template "region" }}
  name          = var.rName
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_scope" "test" {
{{- template "region" }}
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = var.rName
            }
          }
        }
      }
    }
  }
}
