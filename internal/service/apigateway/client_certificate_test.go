package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayClientCertificate_basic(t *testing.T) {
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(resourceName, &conf),
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
				Config: testAccClientCertificateConfig_basic_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/clientcertificates/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", "Hello from TF acceptance test - updated"),
				),
			},
		},
	})
}

func TestAccAPIGatewayClientCertificate_tags(t *testing.T) {
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(resourceName, &conf),
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
				Config: testAccClientCertificateTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClientCertificateTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayClientCertificate_disappears(t *testing.T) {
	var conf apigateway.ClientCertificate
	resourceName := "aws_api_gateway_client_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientCertificateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientCertificateExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceClientCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClientCertificateExists(n string, res *apigateway.ClientCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Client Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetClientCertificateInput{
			ClientCertificateId: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetClientCertificate(req)
		if err != nil {
			return err
		}

		*res = *out

		return nil
	}
}

func testAccCheckClientCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_client_certificate" {
			continue
		}

		req := &apigateway.GetClientCertificateInput{
			ClientCertificateId: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetClientCertificate(req)
		if err == nil {
			return fmt.Errorf("API Gateway Client Certificate still exists: %s", out)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != apigateway.ErrCodeNotFoundException {
			return err
		}

		return nil
	}

	return nil
}

const testAccClientCertificateConfig_basic = `
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"
}
`

const testAccClientCertificateConfig_basic_updated = `
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test - updated"
}
`

func testAccClientCertificateTags1Config(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccClientCertificateTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = "Hello from TF acceptance test"

  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
