// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayDomainNameAccessAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainNameAccessAssociation
	resourceName := "aws_api_gateway_domain_name_access_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameAccessAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAccessAssociationConfig_basic(rName, domain, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameAccessAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "access_association_source", "aws_vpc_endpoint.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "access_association_source_type", "VPCE"),
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

func TestAccAPIGatewayDomainNameAccessAssociation_Identity_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.DomainNameAccessAssociation
	resourceName := "aws_api_gateway_domain_name_access_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameAccessAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAccessAssociationConfig_basic(rName, domain, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameAccessAssociationExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("apigateway", regexache.MustCompile(fmt.Sprintf(`/domainnameaccessassociations/domainname/%s\+[0-9a-z]{10}/vpcesource/vpce-[[:xdigit:]]+`, domain)))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayDomainNameAccessAssociation_Identity_RegionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_api_gateway_domain_name_access_association.test"
	domain := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameAccessAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAccessAssociationConfig_regionOverride(domain, key, certificate),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("apigateway", regexache.MustCompile(fmt.Sprintf(`/domainnameaccessassociations/domainname/%s\+[0-9a-z]{10}/vpcesource/vpce-[[:xdigit:]]+`, domain)))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.CrossRegionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayDomainNameAccessAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.DomainNameAccessAssociation
	resourceName := "aws_api_gateway_domain_name_access_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameAccessAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAccessAssociationConfig_basic(rName, domain, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameAccessAssociationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceDomainNameAccessAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainNameAccessAssociationExists(ctx context.Context, n string, v *types.DomainNameAccessAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindDomainNameAccessAssociationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainNameAccessAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_domain_name_access_association" {
				continue
			}

			_, err := tfapigateway.FindDomainNameAccessAssociationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Domain Name Access Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainNameAccessAssociationConfig_basic(rName, domainName, key, certificate string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_api_gateway_domain_name_access_association" "test" {
  access_association_source      = aws_vpc_endpoint.test.id
  access_association_source_type = "VPCE"
  domain_name_arn                = aws_api_gateway_domain_name.test.arn
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = %[2]q
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}
`, rName, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccDomainNameAccessAssociationConfig_regionOverride(domainName, key, certificate string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn_RegionOverride(acctest.AlternateRegion()),
		fmt.Sprintf(`
resource "aws_api_gateway_domain_name_access_association" "test" {
  region = %[2]q

  access_association_source      = aws_vpc_endpoint.test.id
  access_association_source_type = "VPCE"
  domain_name_arn                = aws_api_gateway_domain_name.test.arn
}

resource "aws_api_gateway_domain_name" "test" {
  region = %[2]q

  domain_name     = %[1]q
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_vpc_endpoint" "test" {
  region = %[2]q

  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

data "aws_region" "current" {
  region = %[2]q
}

resource "aws_vpc" "test" {
  region = %[2]q

  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_default_security_group" "test" {
  region = %[2]q

  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  region = %[2]q

  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_acm_certificate" "test" {
  region = %[2]q

  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}
`, domainName, acctest.AlternateRegion(), acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}
