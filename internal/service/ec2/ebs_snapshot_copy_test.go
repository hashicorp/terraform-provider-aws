// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSSnapshotCopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEBSSnapshotCopy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccEBSSnapshotCopyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEBSSnapshotCopyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Copy Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_regions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshot awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &snapshot),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_storageTier(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSSnapshotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig_storageTier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "archive"),
				),
			},
		},
	})
}

func testAccEBSSnapshotCopyBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotCopyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(rName), `
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
}
`)
}

func testAccEBSSnapshotCopyConfig_storageTier(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
  storage_tier       = "archive"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotCopyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccEBSSnapshotCopyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEBSSnapshotCopyConfig_description(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_copy" "test" {
  description        = "Copy Snapshot Acceptance Test"
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotCopyConfig_regions(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
data "aws_availability_zones" "alternate_available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_ebs_volume" "test" {
  provider          = "awsalternate"
  availability_zone = data.aws_availability_zones.alternate_available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  provider  = "awsalternate"
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.alternate.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotCopyConfig_kms(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
  encrypted          = true
  kms_key_id         = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
