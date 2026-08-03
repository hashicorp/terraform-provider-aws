resource "aws_config_aggregate_authorization" "test" {
  account_id            = data.aws_caller_identity.current.account_id
  authorized_aws_region = data.aws_region.default.name
{{- template "tags" . }}
}

data "aws_caller_identity" "current" {}

data "aws_region" "default" {}
