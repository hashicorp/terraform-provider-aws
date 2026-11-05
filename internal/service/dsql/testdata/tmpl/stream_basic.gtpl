resource "aws_dsql_stream" "test" {
{{- template "region" }}
  cluster_identifier = aws_dsql_cluster.test.identifier
  format             = "JSON"
  ordering           = "UNORDERED"

  target_definition {
    kinesis {
      role_arn   = aws_iam_role.test.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }

  depends_on = [aws_iam_role_policy.test]

{{- template "tags" . }}
}

resource "aws_dsql_cluster" "test" {
{{- template "region" }}
}

data "aws_service_principal" "test" {
{{- template "region" }}
  service_name = "dsql"
}

data "aws_iam_policy_document" "test_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.test.name]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume_role.json
}

data "aws_iam_policy_document" "test_role_policy" {
  statement {
    actions   = ["kinesis:DescribeStreamSummary", "kinesis:ListShards", "kinesis:PutRecord", "kinesis:PutRecords"]
    resources = [aws_kinesis_stream.test.arn]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test_role_policy.json
}

resource "aws_kinesis_stream" "test" {
{{- template "region" }}
  name        = var.rName
  shard_count = 1
}
