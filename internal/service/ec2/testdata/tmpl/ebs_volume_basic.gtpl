resource "aws_ebs_volume" "test" {
{{- template "region" }}
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
{{- template "tags" . }}
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}