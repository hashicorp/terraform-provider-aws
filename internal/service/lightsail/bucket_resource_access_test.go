// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailBucketResourceAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket_resource_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketResourceAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketResourceAccessConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketResourceAccessExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucketName, bucketName),
					resource.TestCheckResourceAttr(resourceName, "resource_name", rName),
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

func TestAccLightsailBucketResourceAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_bucket_resource_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketResourceAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketResourceAccessConfig_basic(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketResourceAccessExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceBucketResourceAccess(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBucketResourceAccessExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		out, err := tflightsail.FindBucketResourceAccessById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("BucketResourceAccess %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketResourceAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_bucket_access_key" {
				continue
			}

			_, err := tflightsail.FindBucketResourceAccessById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResBucketResourceAccess, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccBucketResourceAccessConfig_base(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_bucket" "test" {
  name      = %[1]q
  bundle_id = "small_1_0"
}
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}


`, bucketName)
}

func testAccBucketResourceAccessConfig_basic(rName string, bucketName string) string {
	return acctest.ConfigCompose(testAccBucketResourceAccessConfig_base(bucketName), fmt.Sprintf(`
resource "aws_lightsail_instance" "test" {
  name              = %[1]q
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_bucket_resource_access" "test" {
  bucket_name   = aws_lightsail_bucket.test.id
  resource_name = aws_lightsail_instance.test.id
}


`, rName))
}
