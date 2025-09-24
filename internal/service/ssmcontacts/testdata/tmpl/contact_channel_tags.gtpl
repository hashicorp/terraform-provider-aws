resource "aws_ssmcontacts_contact_channel" "test" {
{{- template "region" }}
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = "test@example.com"
  }

  name = var.rName
  type = "EMAIL"

{{- template "tags" . }}
}

resource "aws_ssmcontacts_contact" "test" {
{{- template "region" }}
  alias = "test-contact-for-${var.rName}"
  type  = "PERSONAL"

  depends_on = [data.aws_ssmincidents_replication_set.test]
{{- template "tags" . }}
}

# testAccContactChannelConfig_base

data "aws_ssmincidents_replication_set" "test" {}

data "aws_region" "current" {
{{- template "region" -}}
}