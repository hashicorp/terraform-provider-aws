resource "aws_odb_exascale_db_storage_vault" "test" {
{{- template "region" }}
  availability_zone_id                             = local.availability_zone_id
  display_name                                     = "ofake-${var.rName}"
  high_capacity_database_storage_total_size_in_gbs = 300
{{- template "tags" . }}
}

data "aws_region" "current" {
{{- template "region" }}
}

locals {
  availability_zone_ids = {
    "eu-west-1" = "euw1-az1"
    "us-east-1" = "use1-az2"
  }

  availability_zone_id = local.availability_zone_ids[data.aws_region.current.name]
}
