resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name
}

resource "aws_iam_user" "test" {
  name = var.rName
}
