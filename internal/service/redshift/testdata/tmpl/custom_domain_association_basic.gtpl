locals {
  custom_domain_name = "${var.rName}.${trimsuffix(var.ACM_CERTIFICATE_ROOT_DOMAIN, ".")}"
}

resource "aws_redshift_subnet_group" "test" {
{{- template "region" }}
  name       = var.rName
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_redshift_cluster" "test" {
{{- template "region" }}
  cluster_identifier                   = var.rName
  cluster_subnet_group_name            = aws_redshift_subnet_group.test.name
  database_name                        = "mydb"
  master_username                      = "foo_test"
  master_password                      = "Mustbe8characters"
  node_type                            = "ra3.large"
  automated_snapshot_retention_period  = 1
  allow_version_upgrade                = false
  skip_final_snapshot                  = true
  availability_zone_relocation_enabled = true
  publicly_accessible                  = false
}

data "aws_route53_zone" "test" {
  name         = var.ACM_CERTIFICATE_ROOT_DOMAIN
  private_zone = false
}

resource "aws_acm_certificate" "test" {
{{- template "region" }}
  domain_name       = local.custom_domain_name
  validation_method = "DNS"
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
{{- template "region" }}
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

resource "aws_redshift_custom_domain_association" "test" {
{{- template "region" }}
  cluster_identifier            = aws_redshift_cluster.test.cluster_identifier
  custom_domain_name            = local.custom_domain_name
  custom_domain_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
