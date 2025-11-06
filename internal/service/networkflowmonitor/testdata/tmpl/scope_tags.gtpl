data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Test scope for ${var.rName}
resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
{{- template "tags" . }}
}