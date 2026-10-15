resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_user" "test" {
  name = var.rName
}

resource "aws_iam_policy" "test" {
  name        = var.rName
  description = "A test policy"

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "iam:ChangePassword"
    ]
    resources = [
      "*"
    ]
  }
}