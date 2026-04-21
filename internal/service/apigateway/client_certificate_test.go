// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayClientCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetClientCertificateOutput
	resourceName := "aws_api_gateway_client_certificate.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/clientcertificates/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Hello from TF acceptance test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientCertificateConfig_basicUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/clientcertificates/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Hello from TF acceptance test - updated"),
				),
			},
		},
	})
}

func TestAccAPIGatewayClientCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetClientCertificateOutput
	resourceName := "aws_api_gateway_client_certificate.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapigateway.ResourceClientCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClientCertificateExists(ctx context.Context, t *testing.T, n string, v *apigateway.GetClientCertificateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		output, err := tfapigateway.FindClientCertificateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClientCertificateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_client_certificate" {
				continue
			}

			_, err := tfapigateway.FindClientCertificateByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Client Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccClientCertificateConfig_basic = `
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"
}
`

const testAccClientCertificateConfig_basicUpdated = `
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test - updated"
}
`
