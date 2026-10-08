resource "aws_config_delivery_channel" "test" {
{{- template "region" }}
  name           = var.rName
  s3_bucket_name = aws_s3_bucket.test.bucket
  depends_on     = [aws_config_configuration_recorder.test]
}

resource "aws_config_configuration_recorder" "test" {
{{- template "region" }}
  name     = var.rName
  role_arn = aws_iam_role.test.arn
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
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "${aws_kms_key.test.arn}"
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = var.rName
  force_destroy = true
}

resource "aws_kms_key" "test" {
{{- template "region" }}
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
