resource "aws_mwaaserverless_workflow" "test" {
{{- template "region" }}
  name     = var.rName
  role_arn = aws_iam_role.test.arn

  definition_s3_location {
    bucket     = aws_s3_bucket.test.id
    object_key = aws_s3_object.test.key
  }
{{- template "tags" . }}
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
{{- template "region" }}
  bucket = aws_s3_bucket.test.id
  key    = "workflow.yaml"

  content = <<-YAML
    name: ${var.rName}
    schedule: "@daily"
    tasks:
      - id: start
        operator: airflow.operators.empty.EmptyOperator
  YAML
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "airflow-serverless.amazonaws.com"
      }
    }]
  })
}
