resource "aws_imagebuilder_container_recipe" "test" {
{{- template "region" }}
  name           = var.rName
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = "FROM $${imagebuilder:parentImage}\n$${imagebuilder:environments}\n$${imagebuilder:components}"

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
{{- template "tags" }}
}

resource "aws_ecr_repository" "test" {
{{- template "region" }}
  name = var.rName
}

data "aws_partition" "current" {}
data "aws_region" "current" {
{{- template "region" }}
}
