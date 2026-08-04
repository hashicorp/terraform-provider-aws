resource "aws_xray_resource_policy" "test" {
{{- template "region" }}
  policy_name                 = var.rName
  bypass_policy_lockout_check = true

  policy_document = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowXRayAccess",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "xray:*",
        "xray:PutResourcePolicy"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
