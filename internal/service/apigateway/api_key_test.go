package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayAPIKey_basic(t *testing.T) {
	var apiKey1 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apikeys/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrSet(resourceName, "value"),
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

func TestAccAPIGatewayAPIKey_tags(t *testing.T) {
	var apiKey1 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAPIKeyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAPIKeyConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccAPIGatewayAPIKey_description(t *testing.T) {
	var apiKey1, apiKey2 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIKeyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey2),
					testAccCheckAPIKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKey_enabled(t *testing.T) {
	var apiKey1, apiKey2 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIKeyConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey2),
					testAccCheckAPIKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKey_value(t *testing.T) {
	var apiKey1 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_value(rName, `MyCustomToken#@&\"'(§!ç)-_*$€¨^£%ù+=/:.;?,|`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "value", `MyCustomToken#@&\"'(§!ç)-_*$€¨^£%ù+=/:.;?,|`),
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

func TestAccAPIGatewayAPIKey_disappears(t *testing.T) {
	var apiKey1 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey1),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceAPIKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAPIKeyExists(n string, res *apigateway.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ApiKey ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetApiKeyInput{
			ApiKey: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetApiKey(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway ApiKey not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAPIKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_api_key" {
			continue
		}

		describe, err := conn.GetApiKeys(&apigateway.GetApiKeysInput{})

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway ApiKey still exists")
			}
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

func testAccCheckAPIKeyNotRecreated(i, j *apigateway.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedDate).Equal(aws.TimeValue(j.CreatedDate)) {
			return fmt.Errorf("API Gateway API Key recreated")
		}

		return nil
	}
}

func testAccAPIKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAPIKeyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAPIKeyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAPIKeyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccAPIKeyConfig_enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  enabled = %[2]t
  name    = %[1]q
}
`, rName, enabled)
}

func testAccAPIKeyConfig_value(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name  = %[1]q
  value = %[2]q
}
`, rName, value)
}
