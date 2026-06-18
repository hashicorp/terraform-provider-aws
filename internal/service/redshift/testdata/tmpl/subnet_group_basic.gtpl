resource "aws_redshift_subnet_group" "test" {
{{- template "region" }}
  name       = var.rName
  subnet_ids = aws_subnet.test[*].id
{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
