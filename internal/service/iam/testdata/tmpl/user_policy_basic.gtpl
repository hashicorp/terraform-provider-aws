resource "aws_iam_user_policy" "test" {
  {{- template "region" }}
  name = var.rName
  user = aws_iam_user.test.name

  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_user" "test" {
  {{- template "region" }}
  name = var.rName
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
