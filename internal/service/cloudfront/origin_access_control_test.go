package cloudfront_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontOriginAccessControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var originaccesscontrol cloudfront.OriginAccessControl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessControlExists(ctx, resourceName, &originaccesscontrol),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrWith(resourceName, "id", func(value string) error {
						if value == "" {
							return fmt.Errorf("expected attribute to be set")
						}

						if id := aws.StringValue(originaccesscontrol.Id); value != id {
							return fmt.Errorf("expected attribute to be equal to %s", id)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "origin_access_control_origin_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "always"),
					resource.TestCheckResourceAttr(resourceName, "signing_protocol", "sigv4"),
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

func TestAccCloudFrontOriginAccessControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var originaccesscontrol cloudfront.OriginAccessControl
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOriginAccessControlExists(ctx, resourceName, &originaccesscontrol),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceOriginAccessControl(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_Name(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_Description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_description(rName, "Acceptance Test 1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Acceptance Test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_description(rName, "Acceptance Test 2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Acceptance Test 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_description(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccCloudFrontOriginAccessControl_SigningBehavior(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_origin_access_control.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOriginAccessControlDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOriginAccessControlConfig_signingBehavior(rName, "never"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "never"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOriginAccessControlConfig_signingBehavior(rName, "no-override"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "signing_behavior", "no-override"),
				),
			},
		},
	})
}

func testAccCheckOriginAccessControlDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_origin_access_control" {
				continue
			}

			_, err := conn.GetOriginAccessControlWithContext(ctx, &cloudfront.GetOriginAccessControlInput{
				Id: aws.String(rs.Primary.ID),
			})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchOriginAccessControl) {
					return nil
				}
				return err
			}

			return create.Error(names.CloudFront, create.ErrActionCheckingDestroyed, tfcloudfront.ResNameOriginAccessControl, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOriginAccessControlExists(ctx context.Context, name string, originaccesscontrol *cloudfront.OriginAccessControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameOriginAccessControl, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameOriginAccessControl, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn()

		resp, err := conn.GetOriginAccessControlWithContext(ctx, &cloudfront.GetOriginAccessControlInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CloudFront, create.ErrActionCheckingExistence, tfcloudfront.ResNameOriginAccessControl, rs.Primary.ID, err)
		}

		*originaccesscontrol = *resp.OriginAccessControl

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn()

	input := &cloudfront.ListOriginAccessControlsInput{}
	_, err := conn.ListOriginAccessControlsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOriginAccessControlConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName)
}

func testAccOriginAccessControlConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName)
}

func testAccOriginAccessControlConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  description                       = %[2]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}
`, rName, description)
}

func testAccOriginAccessControlConfig_signingBehavior(rName, signingBehavior string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_access_control" "test" {
  name                              = %[1]q
  origin_access_control_origin_type = "s3"
  signing_behavior                  = %[2]q
  signing_protocol                  = "sigv4"
}
`, rName, signingBehavior)
}
