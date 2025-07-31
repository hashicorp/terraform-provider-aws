resource "aws_imagebuilder_image_recipe" "test" {
{{- template "region" }}
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = var.rName
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
{{- template "tags" }}
}

resource "aws_imagebuilder_component" "test" {
{{- template "region" }}
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = var.rName
  platform = "Linux"
  version  = "1.0.0"
}

data "aws_partition" "current" {}
data "aws_region" "current" {
{{- template "region" }}
}
