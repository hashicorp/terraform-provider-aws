resource "aws_bedrock_model_invocation_job" "test" {
{{- template "region" }}
  job_name = var.rName
  model_id = "us.amazon.nova-2-lite-v1:0"
  role_arn = aws_iam_role.test.arn

  input_data_config {
    s3_input_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/input/"
    }
  }

  output_data_config {
    s3_output_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]

}

# testAccModelInvocationJobConfig_base

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {
{{- template "region" -}}
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "input" {
{{- template "region" }}
  bucket  = aws_s3_bucket.test.id
  key     = "input/records.jsonl"
  content = <<-EOT
    {"recordId": "1", "modelInput": {"schemaVersion": "messages-v1", "messages": [{"role": "user", "content": [{"text": "Hello, world!"}]}]}}
  EOT
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "bedrock.amazonaws.com"
      }
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
        ArnLike = {
          "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:model-invocation-job/*"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "S3Access"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject",
          "s3:GetBucketLocation",
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*",
        ]
      },
      {
        Sid    = "CrossRegionInference"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:inference-profile/us.amazon.nova-2-lite-v1:0",
          "arn:${data.aws_partition.current.partition}:bedrock:*::foundation-model/amazon.nova-2-lite-v1:0",
        ]
      }
    ]
  })
}
