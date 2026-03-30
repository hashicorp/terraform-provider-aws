resource "aws_config_configuration_aggregator" "test" {
  name = var.rName

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.region]
  }
{{- template "tags" . }}
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}
