resource "aws_apigatewayv2_vpc_link" "test" {
  name               = var.rName
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id

{{- template "tags" . }}
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
