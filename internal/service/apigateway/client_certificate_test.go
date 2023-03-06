package apigateway_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAPIGatewayClientCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/clientcertificates/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "Hello from TF acceptance test"),
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
					testAccCheckClientCertificateExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/clientcertificates/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "Hello from TF acceptance test - updated"),
				),
			},
		},
	})
}

func TestAccAPIGatewayClientCertificate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientCertificateConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClientCertificateConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayClientCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceClientCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClientCertificateExists(ctx context.Context, n string, v *apigateway.ClientCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Client Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn()

		output, err := tfapigateway.FindClientCertificateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClientCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_client_certificate" {
				continue
			}

			_, err := tfapigateway.FindClientCertificateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccClientCertificateConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccClientCertificateConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
