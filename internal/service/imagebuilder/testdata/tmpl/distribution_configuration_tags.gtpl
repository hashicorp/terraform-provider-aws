resource "aws_imagebuilder_distribution_configuration" "test" {
{{- template "region" }}
  name = var.rName

  distribution {
    ami_distribution_configuration {
      name = "test-name-{{`{{ imagebuilder:buildDate }}`}}"
    }

    region = data.aws_region.current.name
  }
{{- template "tags" }}
}

data "aws_region" "current" {
{{- template "region" }}
}
