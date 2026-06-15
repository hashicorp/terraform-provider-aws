resource "aws_osis_pipeline" "test" {
{{- template "region" }}
  pipeline_name               = var.rName
  pipeline_configuration_body = <<-EOT
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
        EOT
  max_units                   = 1
  min_units                   = 1
{{- template "tags" . }}
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
