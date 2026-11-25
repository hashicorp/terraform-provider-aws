resource "aws_route53_health_check_source" "test" {
  vpc_id = aws_vpc.test.id

  dynamic "subnet_configuration" {
    for_each = aws_subnet.test

    content {
      availability_zone_id = subnet_configuration.value.availability_zone_id
      security_group_ids   = [aws_security_group.test.id]
      subnet_id            = subnet_configuration.value.id
    }
  }
{{- template "tags" . }}
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = var.rName
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 3 }}
