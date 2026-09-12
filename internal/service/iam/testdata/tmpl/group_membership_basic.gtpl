resource "aws_iam_group_membership" "test" {
  {{- template "region" }}
  name  = var.rName
  users = [aws_iam_user.test.name]
  group = aws_iam_group.test.name
}

resource "aws_iam_group" "test" {
  {{- template "region" }}
  name = var.rName
}

resource "aws_iam_user" "test" {
  {{- template "region" }}
  name = var.rName
}
