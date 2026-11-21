resource "aws_eventbridgev2_resource_policy" "test" {
{{- template "region" }}
  resource_arn = aws_eventbridgev2_event_bus.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowAccount"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action   = "events:PutEvents"
      Resource = aws_eventbridgev2_event_bus.test.arn
    }]
  })
}

resource "aws_eventbridgev2_event_bus" "test" {
{{- template "region" }}
  name = var.rName
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}
