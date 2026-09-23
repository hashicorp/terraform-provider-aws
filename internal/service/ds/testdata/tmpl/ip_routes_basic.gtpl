resource "aws_directory_service_ip_routes" "test" {
{{- template "region" }}
  directory_id = aws_directory_service_directory.test.id

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = var.rName
  }
}

resource "aws_directory_service_directory" "test" {
{{- template "region" }}
  name     = "corp.example.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
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

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

data "aws_availability_zones" "available" {
{{- template "region" }}
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
