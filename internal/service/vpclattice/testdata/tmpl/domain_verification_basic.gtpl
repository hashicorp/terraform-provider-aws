resource "aws_vpclattice_domain_verification" "test" {
  domain_name = "${var.rName}.example.com"

  {{- template "tags" . }}
}
