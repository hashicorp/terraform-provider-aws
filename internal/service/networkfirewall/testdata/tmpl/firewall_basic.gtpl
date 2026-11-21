resource "aws_networkfirewall_firewall" "test" {
{{- template "region" }}
  name                = var.rName
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
{{- template "tags" . }}
}

resource "aws_networkfirewall_firewall_policy" "test" {
{{- template "region" }}
  name = var.rName

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/24"
}
