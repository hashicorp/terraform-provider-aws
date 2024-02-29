// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessCustomDomainAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customDomainAssociation redshiftserverless.GetCustomDomainAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshiftserverless_custom_domain_association.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, resourceName, &customDomainAssociation),
					resource.TestCheckResourceAttrSet("aws_redshiftserverless_custom_domain_association.test", "custom_domain_certificate_expiry_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s,%s", rName, domain),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRedshiftServerlessCustomDomainAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customDomainAssociation redshiftserverless.GetCustomDomainAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshiftserverless_custom_domain_association.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, resourceName, &customDomainAssociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshiftserverless.ResourceCustomDomainAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomDomainAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshiftserverless_custom_domain_association" {
				continue
			}

			_, err := conn.GetCustomDomainAssociation(ctx, &redshiftserverless.GetCustomDomainAssociationInput{
				CustomDomainName: aws.String(rs.Primary.Attributes["custom_domain_name"]),
				WorkgroupName:    aws.String(rs.Primary.Attributes["workgroup_name"]),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(
					names.RedshiftServerless,
					create.ErrActionCheckingDestroyed,
					tfredshiftserverless.ResNameCustomDomainAssociation,
					fmt.Sprintf("%s,%s", rs.Primary.Attributes["workgroup_name"], rs.Primary.Attributes["custom_domain_name"]),
					err)
			}

			return create.Error(
				names.RedshiftServerless,
				create.ErrActionCheckingDestroyed,
				tfredshiftserverless.ResNameCustomDomainAssociation,
				fmt.Sprintf("%s,%s", rs.Primary.Attributes["workgroup_name"], rs.Primary.Attributes["custom_domain_name"]),
				errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCustomDomainAssociationExists(ctx context.Context, name string, customdomainassociation *redshiftserverless.GetCustomDomainAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RedshiftServerless, create.ErrActionCheckingExistence, tfredshiftserverless.ResNameCustomDomainAssociation, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)
		resp, err := conn.GetCustomDomainAssociation(ctx, &redshiftserverless.GetCustomDomainAssociationInput{
			CustomDomainName: aws.String(rs.Primary.Attributes["custom_domain_name"]),
			WorkgroupName:    aws.String(rs.Primary.Attributes["workgroup_name"]),
		})

		if err != nil {
			return create.Error(
				names.RedshiftServerless,
				create.ErrActionCheckingExistence,
				tfredshiftserverless.ResNameCustomDomainAssociation,
				fmt.Sprintf("%s,%s", rs.Primary.Attributes["workgroup_name"], rs.Primary.Attributes["custom_domain_name"]),
				err)
		}

		*customdomainassociation = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessClient(ctx)

	input := &redshiftserverless.ListCustomDomainAssociationsInput{}
	_, err := conn.ListCustomDomainAssociations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCustomDomainAssociationConfig_basic(rName, rootDomain, domain string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[3]q
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
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}

resource "aws_redshiftserverless_custom_domain_association" "test" {
  workgroup_name                = aws_redshiftserverless_workgroup.test.workgroup_name
  custom_domain_name            = %[3]q
  custom_domain_certificate_arn = aws_acm_certificate.test.arn
}
`, rName, rootDomain, domain)
}
