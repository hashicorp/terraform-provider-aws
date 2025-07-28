resource "aws_imagebuilder_image" "test" {
{{- template "region" }}
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
{{- template "tags" }}
}

{{ template  "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
}

resource "aws_default_route_table" "test" {
{{- template "region" }}
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_default_security_group" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_internet_gateway" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}

resource "aws_iam_role" "test" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilder" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilder"
  role       = aws_iam_role.test.name
}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilder,
  ]
}

resource "aws_imagebuilder_image_recipe" "test" {
{{- template "region" }}
  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
  }

  name         = var.rName
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
{{- template "region" }}
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = var.rName
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}

data "aws_imagebuilder_component" "update-linux" {
{{- template "region" }}
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.region}:aws:component/update-linux/1.0.2"
}
data "aws_partition" "current" {}
data "aws_region" "current" {
{{- template "region" }}
}

