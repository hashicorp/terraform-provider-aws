data "aws_availability_zones" "available" {
{{- template "region" }}
  exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
{{- template "region" }}
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = var.rName
  }
}

resource "aws_directory_service_directory" "test" {
{{- template "region" }}
  name     = "${var.rName}.test"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}