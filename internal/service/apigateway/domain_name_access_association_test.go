// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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

resource "aws_acm_certificate" "test" {
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = %[2]q
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_api_gateway_domain_name_access_association" "test" {
  access_association_source      = aws_vpc_endpoint.test.id
  access_association_source_type = "VPCE"
  domain_name_arn                = aws_api_gateway_domain_name.test.arn
}
`, rName, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}
