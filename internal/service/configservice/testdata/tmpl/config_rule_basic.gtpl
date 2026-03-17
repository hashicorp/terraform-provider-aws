resource "aws_config_config_rule" "test" {
  {{- template "region" }}
  name = var.rName

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.test]
{{- template "tags" . }}
}

# testAccConfigRuleConfig_base

data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  name     = var.rName
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.test.name
}
