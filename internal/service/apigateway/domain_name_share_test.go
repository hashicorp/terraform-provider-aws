// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayDomainNameShare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name_share.test"
	rName := acctest.RandomSubdomain(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameShareConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameShareExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "allowed_accounts.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "allowed_accounts.0", "data.aws_caller_identity.current", names.AttrAccountID),
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

func TestAccAPIGatewayDomainNameShare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name_share.test"
	rName := acctest.RandomSubdomain(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameShareConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameShareExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapigateway.ResourceDomainNameShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainNameShareExists(ctx context.Context, t *testing.T, n string, v *apigateway.GetDomainNameOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		output, err := tfapigateway.FindDomainNameShareByDomainNameID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainNameShareDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_domain_name_share" {
				continue
			}

			_, err := tfapigateway.FindDomainNameShareByDomainNameID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Domain Name Share %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainNameShareConfig_basic(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name_share" "test" {
  domain_name_id    = aws_api_gateway_domain_name.test.domain_name_id
  allowed_accounts  = [data.aws_caller_identity.current.account_id]
}

data "aws_caller_identity" "current" {}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = %[1]q
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}
