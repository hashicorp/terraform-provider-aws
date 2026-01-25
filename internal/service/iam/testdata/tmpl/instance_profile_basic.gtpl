resource "aws_iam_instance_profile" "test" {
  name = var.rName
  role = aws_iam_role.test.name
{{- template "tags" . }}
}

resource "aws_iam_role" "test" {
  name = "${var.rName}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}
