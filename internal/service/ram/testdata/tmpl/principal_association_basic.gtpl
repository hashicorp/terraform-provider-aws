resource "aws_ram_principal_association" "test" {
{{- template "region" }}
  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_resource_share" "test" {
{{- template "region" }}
  allow_external_principals = true
  name                      = var.rName
}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}

{{ template "acctest.ConfigAlternateAccountProvider" }}
