
resource "aws_cloudwatch_event_bus_policy" "test" {
{{- template "region" }}
  policy         = data.aws_iam_policy_document.access.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}

resource "aws_cloudwatch_event_bus" "test" {
{{- template "region" }}
  name = var.rName
}

data "aws_iam_policy_document" "access" {
  statement {
    sid    = "test-resource-policy"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutEvents",
      "events:PutRule"
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
}
