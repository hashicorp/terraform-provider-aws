resource "aws_iam_group_policy_attachment" "test" {
  {{- template "region" }}
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_group" "test" {
  {{- template "region" }}
  name = var.rName
}

resource "aws_iam_policy" "test" {
  {{- template "region" }}
  name = var.rName

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  {{- template "region" }}
  statement {
    effect = "Allow"
    actions = [
      "sts:GetCallerIdentity"
    ]
    resources = [
      "*"
    ]
  }
}
