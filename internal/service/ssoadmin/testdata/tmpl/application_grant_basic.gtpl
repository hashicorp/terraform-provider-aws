data "aws_ssoadmin_instances" "test" {
{{- template "region" -}}
}

resource "aws_ssoadmin_application" "test" {
{{- template "region" }}
  name                     = var.rName
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_application_grant" "test" {
{{- template "region" }}
  application_arn = aws_ssoadmin_application.test.application_arn
  grant_type      = "authorization_code"

  grant {
    authorization_code {
      redirect_uris = ["https://example.com/callback"]
    }
  }
}
