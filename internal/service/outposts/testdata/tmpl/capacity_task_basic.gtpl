data "aws_outposts_outposts" "test" {
{{- template "region" }}
}

data "aws_outposts_outpost_instance_types" "test" {
{{- template "region" }}
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_outposts_capacity_task" "test" {
{{- template "region" }}
  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }
}
