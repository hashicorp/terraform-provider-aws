package cloudfront_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
)

func TestAccCloudFrontPublicKey_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudfront_public_key.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_public_key.example", "comment", "test key"),
					resource.TestMatchResourceAttr("aws_cloudfront_public_key.example", "caller_reference", regexp.MustCompile(fmt.Sprintf("^%s", resource.UniqueIdPrefix))),
					resource.TestCheckResourceAttr("aws_cloudfront_public_key.example", "name", fmt.Sprintf("tf-acc-test-%d", rInt)),
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

func TestAccCloudFrontPublicKey_disappears(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudfront_public_key.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExistence(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourcePublicKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontPublicKey_namePrefix(t *testing.T) {
	startsWithPrefix := regexp.MustCompile("^tf-acc-test-")
	resourceName := "aws_cloudfront_public_key.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExistence(resourceName),
					resource.TestMatchResourceAttr("aws_cloudfront_public_key.example", "name", startsWithPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccCloudFrontPublicKey_update(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_cloudfront_public_key.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicKeyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_public_key.example", "comment", "test key"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPublicKeyUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicKeyExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_public_key.example", "comment", "test key1"),
				),
			},
		},
	})
}

func testAccCheckPublicKeyExistence(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		params := &cloudfront.GetPublicKeyInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetPublicKey(params)
		if err != nil {
			return fmt.Errorf("Error retrieving CloudFront PublicKey: %s", err)
		}
		return nil
	}
}

func testAccCheckPublicKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_public_key" {
			continue
		}

		params := &cloudfront.GetPublicKeyInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetPublicKey(params)
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchPublicKey) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("CloudFront PublicKey (%s) was not deleted", rs.Primary.ID)
	}

	return nil
}

func testAccPublicKeyConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "example" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = "tf-acc-test-%d"
}
`, rInt)
}

func testAccPublicKeyConfig_namePrefix() string {
	return `
resource "aws_cloudfront_public_key" "example" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name_prefix = "tf-acc-test-"
}
`
}

func testAccPublicKeyUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "example" {
  comment     = "test key1"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = "tf-acc-test-%d"
}
`, rInt)
}
