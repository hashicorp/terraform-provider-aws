// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "standard"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccEC2EBSSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEBSSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSSnapshot_storageTier(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_storageTier(rName, "archive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "archive"),
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

func TestAccEC2EBSSnapshot_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, names.AttrARN),
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

func TestAccEC2EBSSnapshot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSSnapshotConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEBSSnapshotConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshot_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
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

func TestAccEC2EBSSnapshot_withKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
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

func testAccCheckSnapshotExists(ctx context.Context, n string, v *awstypes.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EBS Snapshot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSnapshotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEBSSnapshotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ebs_snapshot" {
				continue
			}

			_, err := tfec2.FindSnapshotByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EBS Snapshot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEBSSnapshotBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), `
resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}
`)
}

func testAccEBSSnapshotConfig_storageTier(rName, tier string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id    = aws_ebs_volume.test.id
  storage_tier = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, tier))
}

func testAccEBSSnapshotConfig_outpost(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ebs_snapshot" "test" {
  volume_id   = aws_ebs_volume.test.id
  outpost_arn = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccEBSSnapshotConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEBSSnapshotConfig_description(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id   = aws_ebs_volume.test.id
  description = "test description"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotConfig_kms(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
  kms_key_id        = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
