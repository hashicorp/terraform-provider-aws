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

func TestAccS3FilesMountTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetMountTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
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

func TestAccS3FilesMountTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetMountTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3files.ResourceMountTarget, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMountTargetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_mount_target" {
				continue
			}
			_, err := tfs3files.FindMountTargetByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("S3 Files Mount Target %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckMountTargetExists(ctx context.Context, t *testing.T, n string, v *s3files.GetMountTargetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		output, err := tfs3files.FindMountTargetByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		*v = *output
		return nil
	}
}

func testAccMountTargetConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_basic(rName),
		acctest.ConfigVPCWithSubnets(rName, 1),
		`
resource "aws_s3files_mount_target" "test" {
  file_system_id = aws_s3files_file_system.test.file_system_id
  subnet_id      = aws_subnet.test[0].id
}
`)
}

func TestAccS3FilesMountTarget_securityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetMountTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_mount_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMountTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig_securityGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMountTargetConfig_securityGroupsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMountTargetExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
				),
			},
		},
	})
}

func testAccMountTargetConfig_securityGroups(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_basic(rName),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id
}

resource "aws_s3files_mount_target" "test" {
  file_system_id  = aws_s3files_file_system.test.file_system_id
  subnet_id       = aws_subnet.test[0].id
  security_groups = [aws_security_group.test1.id]
}
`, rName))
}

func testAccMountTargetConfig_securityGroupsUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_basic(rName),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id
}

resource "aws_s3files_mount_target" "test" {
  file_system_id  = aws_s3files_file_system.test.file_system_id
  subnet_id       = aws_subnet.test[0].id
  security_groups = [aws_security_group.test1.id, aws_security_group.test2.id]
}
`, rName))
}
