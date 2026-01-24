resource "aws_codeartifact_repository_permissions_policy" "test" {
{{- template "region" }}
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}

resource "aws_codeartifact_repository" "test" {
{{- template "region" }}
  repository = var.rName
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_domain" "test" {
{{- template "region" }}
  domain         = var.rName
  encryption_key = aws_kms_key.test.arn
}

resource "aws_kms_key" "test" {
{{- template "region" }}
  description             = var.rName
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
