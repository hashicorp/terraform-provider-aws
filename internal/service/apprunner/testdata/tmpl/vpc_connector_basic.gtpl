resource "aws_apprunner_vpc_connector" "test" {
{{- template "region" }}
  vpc_connector_name = var.rName
  subnets            = aws_subnet.test[*].id
  security_groups    = [aws_security_group.test.id]

{{- template "tags" . }}
}

# testAccVPCConnectorConfig_base

resource "aws_security_group" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
  name   = var.rName
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
