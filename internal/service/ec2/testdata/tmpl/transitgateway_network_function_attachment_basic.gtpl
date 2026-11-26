resource "aws_ec2_transit_gateway_network_function_attachment" "test" {
{{- template "region" }}
  availability_zone_ids = [aws_subnet.test[0].availability_zone_id]
  service_name          = aws_vpc_endpoint_service.test.service_name
  transit_gateway_id    = aws_ec2_transit_gateway.test.id

{{- template "tags" . }}
}

resource "aws_ec2_transit_gateway" "test" {
{{- template "region" }}
  tags = {
    Name = var.rName
  }
}

resource "aws_vpc_endpoint_service" "test" {
{{- template "region" }}
  acceptance_required        = false
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = var.rName
  }
}

resource "aws_lb" "test" {
{{- template "region" }}
  load_balancer_type = "gateway"
  name               = var.rName
  subnets            = aws_subnet.test[*].id
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
