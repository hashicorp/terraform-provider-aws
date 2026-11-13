resource "aws_ebs_volume_copy" "test" {
{{- template "region" }}
  source_volume_id = aws_ebs_volume.test.id
}

resource "aws_ebs_volume" "test" {
{{- template "region" }}
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
