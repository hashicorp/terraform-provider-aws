resource "aws_ssmcontacts_contact_channel" "test" {
{{- template "region" }}
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = "test@example.com"
  }

  name = var.rName
  type = "EMAIL"
}

resource "aws_ssmcontacts_contact" "test" {
{{- template "region" }}
  alias = "test-contact-for-${var.rName}"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

# testAccContactChannelConfig_base

resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = data.aws_region.current.region
  }
}

data "aws_region" "current" {
{{- template "region" -}}
}
