resource "aws_organizations_policy" "test" {
  name = var.rName

  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
{{- template "tags" . }}
}
