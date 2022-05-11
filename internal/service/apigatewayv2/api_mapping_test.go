package apigatewayv2_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// These tests need to be serialized, else resources get orphaned after "TooManyRequests" errors.
func TestAccAPIGatewayV2APIMapping_basic(t *testing.T) {
	var certificateArn string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// Create an ACM certificate to be used by all the tests.
	// It is created outside the Terraform configurations because deletion
	// of CloudFront distribution backing the API Gateway domain name is asynchronous
	// and can take up to 60 minutes and the distribution keeps the certificate alive.
	t.Run("createCertificate", func(t *testing.T) {
		testAccAPIMapping_createCertificate(t, rName, &certificateArn)
	})

	testCases := map[string]func(t *testing.T, rName string, certificateArn *string){
		"basic":         testAccAPIMapping_basic,
		"disappears":    testAccAPIMapping_disappears,
		"ApiMappingKey": testAccAPIMapping_ApiMappingKey,
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t, rName, &certificateArn)
		})
	}
}

func testAccAPIMapping_createCertificate(t *testing.T, rName string, certificateArn *string) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: "# Dummy config.",
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingCreateCertificate(rName, certificateArn),
				),
			},
		},
	})

	log.Printf("[INFO] Created ACM certificate %s", *certificateArn)
}

func testAccAPIMapping_basic(t *testing.T, rName string, certificateArn *string) {
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"
	domainNameResourceName := "aws_apigatewayv2_domain_name.test"
	stageResourceName := "aws_apigatewayv2_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIMappingConfig_basic(rName, *certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(resourceName, &domainName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAPIMappingImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAPIMapping_disappears(t *testing.T, rName string, certificateArn *string) {
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIMappingConfig_basic(rName, *certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(resourceName, &domainName, &v),
					testAccCheckAPIMappingDisappears(&domainName, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAPIMapping_ApiMappingKey(t *testing.T, rName string, certificateArn *string) {
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"
	domainNameResourceName := "aws_apigatewayv2_domain_name.test"
	stageResourceName := "aws_apigatewayv2_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIMappingConfig_apiMappingKey(rName, *certificateArn, "$context.domainName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(resourceName, &domainName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.domainName"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				Config: testAccAPIMappingConfig_apiMappingKey(rName, *certificateArn, "$context.apiId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(resourceName, &domainName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.apiId"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", domainNameResourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "stage", stageResourceName, "name")),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAPIMappingImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAPIMappingCreateCertificate(rName string, certificateArn *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		privateKey := acctest.TLSRSAPrivateKeyPEM(2048)
		certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(privateKey, fmt.Sprintf("%s.example.com", rName))

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

		output, err := conn.ImportCertificate(&acm.ImportCertificateInput{
			Certificate: []byte(certificate),
			PrivateKey:  []byte(privateKey),
			Tags: tfacm.Tags(tftags.New(map[string]interface{}{
				"Name": rName,
			}).IgnoreAWS()),
		})
		if err != nil {
			return err
		}

		*certificateArn = *output.CertificateArn

		return nil
	}
}

func testAccCheckAPIMappingDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_api_mapping" {
			continue
		}

		_, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
			ApiMappingId: aws.String(rs.Primary.ID),
			DomainName:   aws.String(rs.Primary.Attributes["domain_name"]),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 API mapping %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAPIMappingDisappears(domainName *string, v *apigatewayv2.GetApiMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		_, err := conn.DeleteApiMapping(&apigatewayv2.DeleteApiMappingInput{
			ApiMappingId: v.ApiMappingId,
			DomainName:   domainName,
		})

		return err
	}
}

func testAccCheckAPIMappingExists(n string, vDomainName *string, v *apigatewayv2.GetApiMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API mapping ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		domainName := aws.String(rs.Primary.Attributes["domain_name"])
		resp, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
			ApiMappingId: aws.String(rs.Primary.ID),
			DomainName:   domainName,
		})
		if err != nil {
			return err
		}

		*vDomainName = *domainName
		*v = *resp

		return nil
	}
}

func testAccAPIMappingImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.ID, rs.Primary.Attributes["domain_name"]), nil
	}
}

func testAccAPIMappingConfig_base(rName, certificateArn string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = %[2]q
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`, rName, certificateArn)
}

func testAccAPIMappingConfig_basic(rName, certificateArn string) string {
	return testAccAPIMappingConfig_base(rName, certificateArn) + testAccStageConfig_basicWebSocket(rName) + `
resource "aws_apigatewayv2_api_mapping" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  domain_name = aws_apigatewayv2_domain_name.test.id
  stage       = aws_apigatewayv2_stage.test.id
}
`
}

func testAccAPIMappingConfig_apiMappingKey(rName, certificateArn, apiMappingKey string) string {
	return testAccAPIMappingConfig_base(rName, certificateArn) + testAccStageConfig_basicWebSocket(rName) + fmt.Sprintf(`
resource "aws_apigatewayv2_api_mapping" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  domain_name = aws_apigatewayv2_domain_name.test.id
  stage       = aws_apigatewayv2_stage.test.id

  api_mapping_key = %[1]q
}
`, apiMappingKey)
}
