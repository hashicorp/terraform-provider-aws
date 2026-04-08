// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3files "github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "access_point_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "access_point_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
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

func TestAccS3FilesAccessPoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetAccessPointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPointExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3files.ResourceAccessPoint, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_access_point" {
				continue
			}
			_, err := tfs3files.FindAccessPointByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("S3 Files Access Point %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckAccessPointExists(ctx context.Context, t *testing.T, n string, v *s3files.GetAccessPointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		output, err := tfs3files.FindAccessPointByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		*v = *output
		return nil
	}
}

func testAccAccessPointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_basic(rName), `
resource "aws_s3files_access_point" "test" {
  file_system_id = aws_s3files_file_system.test.file_system_id
}
`)
}
