resource "aws_s3tables_table_replication" "test" {
{{- template "region" }}
  table_arn = aws_s3tables_table.test.arn
  role      = aws_iam_role.test.arn

  rule {
    destination {
      destination_table_bucket_arn = aws_s3tables_table_bucket.target.arn
    }
  }
}

resource "aws_s3tables_table" "test" {
{{- template "region" }}
  name             = replace(var.rName, "-", "_")
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
{{- template "region" }}
  namespace        = replace(var.rName, "-", "_")
  table_bucket_arn = aws_s3tables_table_bucket.source.arn

  lifecycle {
    create_before_destroy = true
  }
}

data "aws_service_principal" "current" {
  service_name = "s3"
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "${data.aws_service_principal.current.name}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3tables_table_bucket" "source" {
{{- template "region" }}
  name = format("%[1]s-source", var.rName)
}

resource "aws_s3tables_table_bucket" "target" {
{{- template "region" }}
  name = format("%[1]s-target", var.rName)
}