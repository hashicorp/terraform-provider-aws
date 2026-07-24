// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftCustomDomainAssociation_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_custom_domain_association.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckCustomDomainAssociation(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrClusterIdentifier),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_certificate_arn", "aws_acm_certificate_validation.test1", "certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "custom_domain_name", domain),
					resource.TestCheckResourceAttrSet(resourceName, "custom_domain_certificate_expiry_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftCustomDomainAssociation_certificateARN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_custom_domain_association.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckCustomDomainAssociation(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_certificate_arn", "aws_acm_certificate_validation.test1", "certificate_arn"),
				),
			},
			{
				Config: testAccCustomDomainAssociationConfig_certificateARN(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_certificate_arn", "aws_acm_certificate_validation.test2", "certificate_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftCustomDomainAssociation_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_custom_domain_association.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckCustomDomainAssociation(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfredshift.ResourceCustomDomainAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomDomainAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_custom_domain_association" {
				continue
			}

			_, err := tfredshift.FindCustomDomainAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterIdentifier], rs.Primary.Attributes["custom_domain_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Custom Domain Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomDomainAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Redshift Custom Domain Association ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		_, err := tfredshift.FindCustomDomainAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterIdentifier], rs.Primary.Attributes["custom_domain_name"])

		return err
	}
}

func testAccPreCheckCustomDomainAssociation(ctx context.Context, t *testing.T) {
	acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)

	conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

	_, err := conn.DescribeCustomDomainAssociations(ctx, &redshift.DescribeCustomDomainAssociationsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCustomDomainAssociationConfigBase(rName, rootDomain string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                   = %[1]q
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
  name         = %[2]q
  private_zone = false
}
`, rName, rootDomain))
}

func testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(testAccCustomDomainAssociationConfigBase(rName, rootDomain), fmt.Sprintf(`
resource "aws_acm_certificate" "test1" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test1" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test1.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test1.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test1.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test1" {
  certificate_arn         = aws_acm_certificate.test1.arn
  validation_record_fqdns = [aws_route53_record.test1.fqdn]
}

resource "aws_redshift_custom_domain_association" "test" {
  cluster_identifier          = aws_redshift_cluster.test.cluster_identifier
  custom_domain_name          = %[1]q
  custom_domain_certificate_arn = aws_acm_certificate_validation.test1.certificate_arn
}
`, domain))
}

func testAccCustomDomainAssociationConfig_certificateARN(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(testAccCustomDomainAssociationConfigBase(rName, rootDomain), fmt.Sprintf(`
resource "aws_acm_certificate" "test1" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test1" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test1.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test1.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test1.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test1" {
  certificate_arn         = aws_acm_certificate.test1.arn
  validation_record_fqdns = [aws_route53_record.test1.fqdn]
}

resource "aws_acm_certificate" "test2" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test2" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test2.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test2.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test2.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test2" {
  certificate_arn         = aws_acm_certificate.test2.arn
  validation_record_fqdns = [aws_route53_record.test2.fqdn]
}

resource "aws_redshift_custom_domain_association" "test" {
  cluster_identifier            = aws_redshift_cluster.test.cluster_identifier
  custom_domain_name            = %[1]q
  custom_domain_certificate_arn = aws_acm_certificate_validation.test2.certificate_arn
}
`, domain))
}
