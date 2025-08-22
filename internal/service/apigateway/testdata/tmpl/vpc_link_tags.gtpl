resource "aws_api_gateway_vpc_link" "test" {
  name        = var.rName
  target_arns = [aws_lb.test.arn]

{{- template "tags" . }}
}

resource "aws_lb" "test" {
  name               = var.rName
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
