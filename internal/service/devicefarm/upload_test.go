// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devicefarm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdevicefarm "github.com/hashicorp/terraform-provider-aws/internal/service/devicefarm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeviceFarmUpload_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.Upload
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, "tf-acc-test-updated")
	resourceName := "aws_devicefarm_upload.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUploadDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUploadConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`upload:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "APPIUM_JAVA_TESTNG_TEST_SPEC"),
					resource.TestCheckResourceAttr(resourceName, "category", "PRIVATE"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrURL},
			},
			{
				Config: testAccUploadConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadExists(ctx, t, resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "devicefarm", regexache.MustCompile(`upload:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "APPIUM_JAVA_TESTNG_TEST_SPEC"),
					resource.TestCheckResourceAttr(resourceName, "category", "PRIVATE"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURL),
				),
			},
		},
	})
}

func TestAccDeviceFarmUpload_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.Upload
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_upload.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUploadDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUploadConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadExists(ctx, t, resourceName, &proj),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceUpload(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceUpload(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeviceFarmUpload_disappears_project(t *testing.T) {
	ctx := acctest.Context(t)
	var proj awstypes.Upload
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_devicefarm_upload.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DeviceFarmEndpointID)
			// Currently, DeviceFarm is only supported in us-west-2
			// https://docs.aws.amazon.com/general/latest/gr/devicefarm.html
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DeviceFarmServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUploadDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUploadConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadExists(ctx, t, resourceName, &proj),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceProject(), "aws_devicefarm_project.test"),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdevicefarm.ResourceUpload(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUploadExists(ctx context.Context, t *testing.T, n string, v *awstypes.Upload) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DeviceFarmClient(ctx)
		resp, err := tfdevicefarm.FindUploadByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DeviceFarm Upload not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckUploadDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DeviceFarmClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devicefarm_upload" {
				continue
			}

			// Try to find the resource
			_, err := tfdevicefarm.FindUploadByARN(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DeviceFarm Upload %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUploadConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_project" "test" {
  name = %[1]q
}

resource "aws_devicefarm_upload" "test" {
  name        = %[1]q
  project_arn = aws_devicefarm_project.test.arn
  type        = "APPIUM_JAVA_TESTNG_TEST_SPEC"
}
`, rName)
}
