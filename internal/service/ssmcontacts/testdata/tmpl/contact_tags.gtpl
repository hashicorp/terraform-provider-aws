resource "aws_ssmcontacts_contact" "test" {
  alias = var.rName
  type  = "PERSONAL"

{{- template "tags" . }}

  depends_on = [aws_ssmincidents_replication_set.test]
}

# testAccContactConfig_base

resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = data.aws_region.current.name
  }
}

data "aws_region" "current" {}
