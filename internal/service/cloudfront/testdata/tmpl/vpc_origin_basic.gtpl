resource "aws_cloudfront_vpc_origin" "test" {
  vpc_origin_endpoint_config {
    name                   = var.rName
    arn                    = aws_lb.test.arn
    http_port              = 8080
    https_port             = 8443
    origin_protocol_policy = "http-only"

    origin_ssl_protocols {
      items    = ["TLSv1.2"]
      quantity = 1
    }
  }

{{- template "tags" . }}
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = var.rName
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_lb" "test" {
  name            = var.rName
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = var.rName
  }

  depends_on = [aws_internet_gateway.test]
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
