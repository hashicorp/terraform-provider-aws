resource "aws_bedrock_custom_model" "test" {
{{- template "region" }}
  custom_model_name     = var.rName
  job_name              = var.rName
  base_model_identifier = data.aws_bedrock_foundation_model.test.model_arn
  role_arn              = aws_iam_role.test.arn

  hyperparameters = {
    "epochCount"              = "1"
    "batchSize"               = "1"
    "learningRate"            = "0.005"
    "learningRateWarmupSteps" = "0"
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.output.id}/data/"
  }

  training_data_config {
    s3_uri = "s3://${aws_s3_bucket.training.id}/data/train.jsonl"
  }

{{- template "tags" . }}
}

# testAccCustomModelConfig_base

data "aws_caller_identity" "current" {}
data "aws_region" "current" {
{{- template "region" -}}
}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "training" {
{{- template "region" }}
  bucket = "${var.rName}-training"
}

resource "aws_s3_bucket" "validation" {
{{- template "region" }}
  bucket = "${var.rName}-validation"
}

resource "aws_s3_bucket" "output" {
{{- template "region" }}
  bucket        = "${var.rName}-output"
  force_destroy = true
}

resource "aws_s3_object" "training" {
{{- template "region" }}
  bucket = aws_s3_bucket.training.id
  key    = "data/train.jsonl"
  source = "test-fixtures/train.jsonl"
}

resource "aws_s3_object" "validation" {
{{- template "region" }}
  bucket = aws_s3_bucket.validation.id
  key    = "data/validate.jsonl"
  source = "test-fixtures/validate.jsonl"
}

resource "aws_iam_role" "test" {
  name = var.rName

  # See https://docs.aws.amazon.com/bedrock/latest/userguide/model-customization-iam-role.html#model-customization-iam-role-trust.
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "bedrock.amazonaws.com"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringEquals": {
        "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnEquals": {
        "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:model-customization-job/*"
      }
    }
  }]
}
EOF
}

# See https://docs.aws.amazon.com/bedrock/latest/userguide/model-customization-iam-role.html#model-customization-iam-role-s3.
resource "aws_iam_policy" "training" {
  name = "${var.rName}-training"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Effect" : "Allow",
      "Action" : [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource" : [
        aws_s3_bucket.training.arn,
        "${aws_s3_bucket.training.arn}/*",
        aws_s3_bucket.validation.arn,
        "${aws_s3_bucket.validation.arn}/*"
      ]
    }]
  })
}

resource "aws_iam_policy" "output" {
  name = "${var.rName}-output"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Effect" : "Allow",
      "Action" : [
        "s3:GetObject",
        "s3:PutObject",
        "s3:ListBucket"
      ],
      "Resource" : [
        aws_s3_bucket.output.arn,
        "${aws_s3_bucket.output.arn}/*"
      ]
    }]
  })
}

resource "aws_iam_role_policy_attachment" "training" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.training.arn
}

resource "aws_iam_role_policy_attachment" "output" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.output.arn
}

data "aws_bedrock_foundation_model" "test" {
{{- template "region" }}
  model_id = "amazon.titan-text-express-v1"
}
