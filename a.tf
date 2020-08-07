data "aws_iam_policy_document" "cloudwatch" {
  statement {
    actions = [
      "logs:PutLogEvents",
      "logs:PutLogEventsBatch",
      "logs:CreateLogStream",
    ]
    effect = "Allow"
    principals {
      type = "Service"
      identifiers = ["es.amazonaws.com"]
    }
    resources = [
    # for k, v in aws_cloudwatch_log_group.es_logs : "${v.arn}:*" This never works
    for cw_arn in aws_cloudwatch_log_group.es_logs.arn : "${cw_arn}:*" # this almost never works, but seems to have worked once
    # "arn:aws:logs:us-east-1:${data.aws_caller_identity.current.account_id}:log-group:*" This works 100% of the time based on my tests
    ]
  }
}

resource "aws_cloudwatch_log_group" "es_logs" {
  for_each = local.a
  name = "/aws/aes/${each.key}"
}