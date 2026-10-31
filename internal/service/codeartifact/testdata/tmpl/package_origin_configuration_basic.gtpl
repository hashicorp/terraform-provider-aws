resource "aws_codeartifact_package_origin_configuration" "test" {
{{- template "region" }}
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = "pypi"
  package    = var.rName

  restrictions {
    publish  = "BLOCK"
    upstream = "ALLOW"
  }
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
