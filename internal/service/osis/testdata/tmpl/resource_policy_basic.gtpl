resource "aws_osis_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_osis_pipeline.test.pipeline_arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "ingestPolicy",
  "Statement": [{
    "Sid": "AllowIngest",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.current.account_id}"
    },
    "Action": [
      "osis:CreatePipelineEndpoint"
    ],
    "Resource": "${aws_osis_pipeline.test.pipeline_arn}"
  }]
}
EOF
}

data "aws_caller_identity" "current" {}

# testAccPipelineConfig_basic

resource "aws_osis_pipeline" "test" {
{{- template "region" }}
  pipeline_name               = var.rName
  pipeline_configuration_body = <<EOS
            version: "2"
            test-pipeline:
              source:
                http:
                  path: "/test"
              sink:
                - s3:
                    aws:
                      sts_role_arn: "${aws_iam_role.test.arn}"
                      region: "${data.aws_region.current.region}"
                    bucket: "test"
                    threshold:
                      event_collect_timeout: "60s"
                    codec:
                      ndjson:
EOS
  max_units                   = 1
  min_units                   = 1
}

data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "osis-pipelines.amazonaws.com"
        }
      },
    ]
  })
}
