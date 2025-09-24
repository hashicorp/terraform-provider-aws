resource "aws_s3control_bucket" "test" {
{{- template "region" }}
  bucket     = var.rName
  outpost_id = data.aws_outposts_outpost.test.id
{{- template "tags" }}
}

data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}