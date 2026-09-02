// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSVolumeCopy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"
	volumeResourceName := "aws_ebs_volume.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			acctest.PreCheckPartitionHasService(t, names.KMS)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(resourceName, "source_volume_id", volumeResourceName, names.AttrID),
				),
			},
			{
				Config: testAccEBSVolumeCopyConfig_kmsKeyID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_volume_id", volumeResourceName, names.AttrID),
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

func TestAccEC2EBSVolumeCopy_kmsKeyAlias(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"
	volumeResourceName := "aws_ebs_volume.test"
	kmsAliasResourceName := "aws_kms_alias.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			acctest.PreCheckPartitionHasService(t, names.KMS)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_kmsKeyAlias(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsAliasResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_volume_id", volumeResourceName, names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_shared(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	resourceName := "aws_ebs_volume_copy.test"
	sourceVolumeResourceName := "aws_ebs_volume.source"
	targetKMSKeyResourceName := "aws_kms_key.target"
	sharedVolumeDataSourceName := "data.aws_ebs_volume.shared"
	sourceIdentityDataSourceName := "data.aws_caller_identity.source"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			acctest.PreCheckPartitionHasService(t, names.KMS)
			acctest.PreCheckPartitionHasService(t, names.RAM)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, names.KMSServiceID, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckEBSVolumeCopyDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_shared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, targetKMSKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_volume_id", sourceVolumeResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(sharedVolumeDataSourceName, names.AttrARN, sourceVolumeResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(sharedVolumeDataSourceName, names.AttrOwnerID, sourceIdentityDataSourceName, names.AttrAccountID),
				),
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceEBSVolumeCopy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_updateSize(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeCopyConfig_updateSize(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_updateIops(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_iops(3000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeType, "gp3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeCopyConfig_iops(4000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeType, "gp3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "4000"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_updateThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_throughput(125),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeType, "gp3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "125"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeCopyConfig_throughput(150),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeType, "gp3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "150"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_updateVolumeType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeCopyConfig_volumeTypeGP3(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrVolumeType, "gp3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "125"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_volume_copy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeCopyConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeCopyConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEBSVolumeCopyConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSVolumeCopyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2EBSVolumeCopy_invalidConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSVolumeCopyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccEBSVolumeCopyConfig_invalidThroughput(),
				ExpectError: regexache.MustCompile(`Invalid Throughput Configuration`),
			},
			{
				Config:      testAccEBSVolumeCopyConfig_invalidIOPS(),
				ExpectError: regexache.MustCompile(`Invalid IOPS Configuration`),
			},
			{
				Config:      testAccEBSVolumeCopyConfig_missingIOPS(),
				ExpectError: regexache.MustCompile(`Missing IOPS Configuration`),
			},
			{
				Config:      testAccEBSVolumeCopyConfig_invalidKMSKeyID(),
				ExpectError: regexache.MustCompile(`Invalid KMS Key Configuration`),
			},
		},
	})
}

func testAccCheckEBSVolumeCopyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckEBSVolumeCopyDestroyWithClient(ctx, s, acctest.ProviderMeta(ctx, t))
	}
}

func testAccCheckEBSVolumeCopyDestroyWithProvider(ctx context.Context) acctest.TestCheckWithProviderFunc {
	return func(s *terraform.State, provider *schema.Provider) error {
		return testAccCheckEBSVolumeCopyDestroyWithClient(ctx, s, provider.Meta().(*conns.AWSClient))
	}
}

func testAccCheckEBSVolumeCopyDestroyWithClient(ctx context.Context, s *terraform.State, client *conns.AWSClient) error {
	conn := client.EC2Client(ctx)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_volume_copy" {
			continue
		}

		_, err := tfec2.FindEBSVolumeByID(ctx, conn, rs.Primary.ID)
		if retry.NotFound(err) {
			return nil
		}
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "EBS Volume Copy", rs.Primary.ID, err)
		}

		return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "EBS Volume Copy", rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckEBSVolumeCopyExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "EBS Volume Copy", name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "EBS Volume Copy", name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindEBSVolumeByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "EBS Volume Copy", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccEBSVolumeCopyConfigBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
}
`)
}

func testAccEBSVolumeCopyConfig_basic() string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), `
resource "aws_ebs_volume_copy" "test" {
  encrypted        = true
  source_volume_id = aws_ebs_volume.test.id
}
`)
}

func testAccEBSVolumeCopyConfig_kmsKeyID() string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_ebs_volume_copy" "test" {
  encrypted        = true
  kms_key_id       = aws_kms_key.test.arn
  source_volume_id = aws_ebs_volume.test.id
}
`)
}

func testAccEBSVolumeCopyConfig_kmsKeyAlias(rName string) string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_ebs_volume_copy" "test" {
  encrypted        = true
  kms_key_id       = aws_kms_alias.test.arn
  source_volume_id = aws_ebs_volume.test.id
}
`, rName))
}

func testAccEBSVolumeCopyConfig_shared(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_caller_identity" "source" {}

data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "source" {
  statement {
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.source.account_id}:root"]
    }
  }

  statement {
    actions = [
      "kms:CreateGrant",
      "kms:Decrypt",
      "kms:DescribeKey",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:ReEncryptFrom",
    ]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.target.account_id}:root"]
    }
  }
}

resource "aws_kms_key" "source" {
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.source.json
}

resource "aws_kms_key" "target" {
  provider = "awsalternate"

  deletion_window_in_days = 7
}

resource "aws_ebs_volume" "source" {
  availability_zone = data.aws_availability_zones.available.names[0]
  encrypted         = true
  kms_key_id        = aws_kms_key.source.arn
  size              = 1
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = true

  permission_arns = [
    "arn:${data.aws_partition.current.partition}:ram::aws:permission/AWSRAMPermissionEBSVolumeCopyAccess",
  ]
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ebs_volume.source.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.target.account_id
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_share_accepter" "test" {
  provider = "awsalternate"

  share_arn = aws_ram_principal_association.test.resource_share_arn
}

data "aws_ebs_volume" "shared" {
  provider = "awsalternate"

  filter {
    name   = "volume-id"
    values = [aws_ebs_volume.source.id]
  }

  depends_on = [
    aws_ram_resource_association.test,
    aws_ram_resource_share_accepter.test,
  ]
}

resource "aws_ebs_volume_copy" "test" {
  provider = "awsalternate"

  encrypted        = true
  kms_key_id       = aws_kms_key.target.arn
  source_volume_id = data.aws_ebs_volume.shared.volume_id
}
`, rName))
}

func testAccEBSVolumeCopyConfig_updateSize() string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), `
resource "aws_ebs_volume_copy" "test" {
  source_volume_id = aws_ebs_volume.test.id
  size             = 2
}
`)
}

func testAccEBSVolumeCopyConfig_iops(iops int) string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), fmt.Sprintf(`
resource "aws_ebs_volume_copy" "test" {
  source_volume_id = aws_ebs_volume.test.id
  volume_type      = "gp3"
  iops             = %d
  size             = 8
}
`, iops))
}

func testAccEBSVolumeCopyConfig_throughput(throughput int) string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), fmt.Sprintf(`
resource "aws_ebs_volume_copy" "test" {
  source_volume_id = aws_ebs_volume.test.id
  volume_type      = "gp3"
  throughput       = %d
  size             = 1
}
`, throughput))
}

func testAccEBSVolumeCopyConfig_volumeTypeGP3() string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), `
resource "aws_ebs_volume_copy" "test" {
  source_volume_id = aws_ebs_volume.test.id
  volume_type      = "gp3"
}
`)
}

func testAccEBSVolumeCopyConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), fmt.Sprintf(`
resource "aws_ebs_volume_copy" "test" {
  source_volume_id = aws_ebs_volume.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccEBSVolumeCopyConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSVolumeCopyConfigBaseConfig(), fmt.Sprintf(`
resource "aws_ebs_volume_copy" "test" {
  source_volume_id = aws_ebs_volume.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

// Throughput should only be configured for gp3 volume types
func testAccEBSVolumeCopyConfig_invalidThroughput() string {
	return `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_ebs_volume_copy" "test" {
  source_volume_id = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:volume/does-not-exist"
  volume_type      = "io1"
  throughput       = 125
}
`
}

// IOPS should only be set for io1, io2, or gp3 volume types
func testAccEBSVolumeCopyConfig_invalidIOPS() string {
	return `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_ebs_volume_copy" "test" {
  source_volume_id = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:volume/does-not-exist"
  volume_type      = "standard"
  iops             = 3000
}
`
}

// IOPS must be set for io1 or io2 volume types
func testAccEBSVolumeCopyConfig_missingIOPS() string {
	return `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_ebs_volume_copy" "test" {
  source_volume_id = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:volume/does-not-exist"
  volume_type      = "io1"
}
`
}

func testAccEBSVolumeCopyConfig_invalidKMSKeyID() string {
	return `
resource "aws_ebs_volume_copy" "test" {
  encrypted        = false
  kms_key_id       = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  source_volume_id = "vol-1234567890abcdef0"
}
`
}
