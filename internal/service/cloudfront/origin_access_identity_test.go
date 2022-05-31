package cloudfront_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
)

func TestAccCloudFrontOriginAccessIdentity_basic(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExistence(resourceName, &origin),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestMatchResourceAttr(resourceName, "caller_reference", regexp.MustCompile(fmt.Sprintf("^%s", resource.UniqueIdPrefix))),
					resource.TestMatchResourceAttr(resourceName, "s3_canonical_user_id", regexp.MustCompile("^[a-z0-9]+")),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_access_identity_path", regexp.MustCompile("^origin-access-identity/cloudfront/[A-Z0-9]+")),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "iam_arn", regexp.MustCompile(fmt.Sprintf("^arn:%s:iam::cloudfront:user/CloudFront Origin Access Identity [A-Z0-9]+", acctest.Partition()))),
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

func TestAccCloudFrontOriginAccessIdentity_noComment(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityNoCommentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExistence(resourceName, &origin),
					resource.TestMatchResourceAttr(resourceName, "caller_reference", regexp.MustCompile(fmt.Sprintf("^%s", resource.UniqueIdPrefix))),
					resource.TestMatchResourceAttr(resourceName, "s3_canonical_user_id", regexp.MustCompile("^[a-z0-9]+")),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_access_identity_path", regexp.MustCompile("^origin-access-identity/cloudfront/[A-Z0-9]+")),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "iam_arn", regexp.MustCompile(fmt.Sprintf("^arn:%s:iam::cloudfront:user/CloudFront Origin Access Identity [A-Z0-9]+", acctest.Partition()))),
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

func TestAccCloudFrontOriginAccessIdentity_disappears(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessIdentityConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessIdentityExistence(resourceName, &origin),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceOriginAccessIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOriginAccessIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_origin_access_identity" {
			continue
		}

		params := &cloudfront.GetCloudFrontOriginAccessIdentityInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCloudFrontOriginAccessIdentity(params)
		if err == nil {
			return fmt.Errorf("CloudFront origin access identity was not deleted")
		}
	}

	return nil
}

func testAccCheckOriginAccessIdentityExistence(r string, origin *cloudfront.GetCloudFrontOriginAccessIdentityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		params := &cloudfront.GetCloudFrontOriginAccessIdentityInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetCloudFrontOriginAccessIdentity(params)
		if err != nil {
			return fmt.Errorf("Error retrieving CloudFront distribution: %s", err)
		}

		*origin = *resp

		return nil
	}
}

const testAccOriginAccessIdentityConfig = `
resource "aws_cloudfront_origin_access_identity" "test" {
  comment = "some comment"
}
`

const testAccOriginAccessIdentityNoCommentConfig = `
resource "aws_cloudfront_origin_access_identity" "test" {
}
`
