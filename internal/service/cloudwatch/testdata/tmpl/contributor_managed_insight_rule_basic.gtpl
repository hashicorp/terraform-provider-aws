resource "aws_cloudwatch_contributor_managed_insight_rule" "test" {
{{- template "region" }}
  resource_arn  = aws_vpc_endpoint_service.test.arn
  template_name = "VpcEndpointService-NewConnectionsByEndpointId-v1"

{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}

resource "aws_lb" "test" {
{{- template "region" }}
  count = 2

  load_balancer_type = "network"
  name               = "${var.rName}-${count.index}"

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false
}

resource "aws_vpc_endpoint_service" "test" {
{{- template "region" }}
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
}