resource "aws_kinesis_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_kinesis_stream.test.arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writestatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:${data.aws_partition.target.partition}:iam::${data.aws_caller_identity.target.account_id}:root"
    },
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:ListShards",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test.arn}"
  }]
}
EOF
}

resource "aws_kinesis_stream" "test" {
{{- template "region" }}
  name        = var.rName
  shard_count = 2
}

data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

data "aws_partition" "target" {
  provider = "awsalternate"
}

{{ template "acctest.ConfigAlternateAccountProvider" }}
