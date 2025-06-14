provider "aws" {
  region = "eu-central-1"
}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = "quick-test"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_s3tables_namespace" "test" {
  namespace        = "namespace_${replace("quick-test", "-", "_")}_test"
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

resource "aws_s3tables_table_bucket" "test" {
  name = "quick-test"
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions      = ["CREATE_TABLE"]
  principal        = aws_iam_role.test.arn

  database {
	name = aws_s3tables_namespace.test.namespace
	catalog_id	   = "s3tablescatalog/quick-test"
  }
}