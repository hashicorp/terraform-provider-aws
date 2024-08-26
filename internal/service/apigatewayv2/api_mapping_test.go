// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// These tests need to be serialized, else resources get orphaned after "TooManyRequests" errors.
func TestAccAPIGatewayV2APIMapping_serial(t *testing.T) {
	t.Parallel()

	var certificateARN string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// Create an ACM certificate to be used by all the tests.
	// It is created outside the Terraform configurations because deletion
	// of CloudFront distribution backing the API Gateway domain name is asynchronous
	// and can take up to 60 minutes and the distribution keeps the certificate alive.
	t.Run("createCertificate", func(t *testing.T) {
		testAccAPIMapping_createCertificate(t, rName, &certificateARN)
	})

	testCases := map[string]func(t *testing.T, rName string, certificateArn *string){
		acctest.CtBasic:      testAccAPIMapping_basic,
		acctest.CtDisappears: testAccAPIMapping_disappears,
		"ApiMappingKey":      testAccAPIMapping_key,
	}
	for name, tc := range testCases { //nolint:paralleltest
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t, rName, &certificateARN)
		})
	}
}

func testAccAPIMapping_createCertificate(t *testing.T, rName string, certificateARN *string) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: "# Empty config",
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingCreateCertificate(ctx, t, rName, certificateARN),
				),
			},
		},
	})

	log.Printf("[INFO] Created ACM certificate %s", aws.ToString(certificateARN))
}

func testAccAPIMapping_basic(t *testing.T, rName string, certificateARN *string) {
	ctx := acctest.Context(t)
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"
	domainNameResourceName := "aws_apigatewayv2_domain_name.test"
	stageResourceName := "aws_apigatewayv2_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIMappingConfig_basic(rName, *certificateARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(ctx, resourceName, &domainName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, domainNameResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStage, stageResourceName, names.AttrName)),
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

func testAccAPIMapping_disappears(t *testing.T, rName string, certificateARN *string) {
	ctx := acctest.Context(t)
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIMappingConfig_basic(rName, *certificateARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(ctx, resourceName, &domainName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceAPIMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAPIMapping_key(t *testing.T, rName string, certificateARN *string) {
	ctx := acctest.Context(t)
	var domainName string
	var v apigatewayv2.GetApiMappingOutput
	resourceName := "aws_apigatewayv2_api_mapping.test"
	domainNameResourceName := "aws_apigatewayv2_domain_name.test"
	stageResourceName := "aws_apigatewayv2_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIMappingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIMappingConfig_key(rName, *certificateARN, "$context.domainName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(ctx, resourceName, &domainName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.domainName"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, domainNameResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStage, stageResourceName, names.AttrName)),
			},
			{
				Config: testAccAPIMappingConfig_key(rName, *certificateARN, "$context.apiId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIMappingExists(ctx, resourceName, &domainName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_mapping_key", "$context.apiId"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, domainNameResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStage, stageResourceName, names.AttrName)),
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

func testAccCheckAPIMappingCreateCertificate(ctx context.Context, t *testing.T, rName string, certificateARN *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		privateKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
		certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKey, fmt.Sprintf("%s.example.com", rName))

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

		output, err := conn.ImportCertificate(ctx, &acm.ImportCertificateInput{
			Certificate: []byte(certificate),
			PrivateKey:  []byte(privateKey),
			Tags: tfacm.Tags(tftags.New(ctx, map[string]interface{}{
				"Name": rName,
			}).IgnoreAWS()),
		})

		if err != nil {
			return err
		}

		*certificateARN = aws.ToString(output.CertificateArn)

		return nil
	}
}

func testAccCheckAPIMappingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_api_mapping" {
				continue
			}

			_, err := tfapigatewayv2.FindAPIMappingByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 API Mapping %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPIMappingExists(ctx context.Context, n string, vDomainName *string, v *apigatewayv2.GetApiMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API Mapping ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		domainName := rs.Primary.Attributes[names.AttrDomainName]
		output, err := tfapigatewayv2.FindAPIMappingByTwoPartKey(ctx, conn, rs.Primary.ID, domainName)

		if err != nil {
			return err
		}

		*vDomainName = domainName
		*v = *output

		return nil
	}
}

func testAccAPIMappingImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.ID, rs.Primary.Attributes[names.AttrDomainName]), nil
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

func testAccAPIMappingConfig_basic(rName, certificateARN string) string {
	return acctest.ConfigCompose(testAccAPIMappingConfig_base(rName, certificateARN), testAccStageConfig_basicWebSocket(rName), `
resource "aws_apigatewayv2_api_mapping" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  domain_name = aws_apigatewayv2_domain_name.test.id
  stage       = aws_apigatewayv2_stage.test.id
}
`)
}

func testAccAPIMappingConfig_key(rName, certificateARN, apiMappingKey string) string {
	return acctest.ConfigCompose(testAccAPIMappingConfig_base(rName, certificateARN), testAccStageConfig_basicWebSocket(rName), fmt.Sprintf(`
resource "aws_apigatewayv2_api_mapping" "test" {
  api_id      = aws_apigatewayv2_api.test.id
  domain_name = aws_apigatewayv2_domain_name.test.id
  stage       = aws_apigatewayv2_stage.test.id

  api_mapping_key = %[1]q
}
`, apiMappingKey))
}
