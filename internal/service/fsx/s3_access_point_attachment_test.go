// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxS3AccessPointAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.S3AccessPointAttachment
	resourceName := "aws_fsx_s3_access_point_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3AccessPointAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccS3AccessPointAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3AccessPointAttachmentExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("openzfs_configuration"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_access_point"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_access_point_alias"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_access_point_arn"), tfknownvalue.RegionalARNRegexp("s3", regexache.MustCompile(`accesspoint/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.S3AccessPointAttachmentTypeOpenzfs)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{names.AttrPolicy},
			},
		},
	})
}

func TestAccFSxS3AccessPointAttachment_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.S3AccessPointAttachment
	resourceName := "aws_fsx_s3_access_point_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3AccessPointAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccS3AccessPointAttachmentConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3AccessPointAttachmentExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_access_point"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrPolicy: knownvalue.NotNull(),
						}),
					})),
				},
			},
		},
	})
}

func TestAccFSxS3AccessPointAttachment_vpcConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.S3AccessPointAttachment
	resourceName := "aws_fsx_s3_access_point_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3AccessPointAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccS3AccessPointAttachmentConfig_vpcConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3AccessPointAttachmentExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_access_point"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrPolicy:           knownvalue.Null(),
							names.AttrVPCConfiguration: knownvalue.ListSizeExact(1),
						}),
					})),
				},
			},
		},
	})
}

func TestAccFSxS3AccessPointAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.S3AccessPointAttachment
	resourceName := "aws_fsx_s3_access_point_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckS3AccessPointAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccS3AccessPointAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3AccessPointAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tffsx.ResourceS3AccessPointAttachment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckS3AccessPointAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.S3AccessPointAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).FSxClient(ctx)

		output, err := tffsx.FindS3AccessPointAttachmentByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckS3AccessPointAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_s3_access_point_attachment" {
				continue
			}

			_, err := tffsx.FindS3AccessPointAttachmentByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx S3 Access Point Attachment %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccS3AccessPointAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_HA_2"
  throughput_capacity = 320
  skip_final_backup   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}
`, rName))
}

func testAccS3AccessPointAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccS3AccessPointAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_s3_access_point_attachment" "test" {
  name = %[1]q
  type = "OPENZFS"

  openzfs_configuration {
    volume_id = aws_fsx_openzfs_volume.test.id

    file_system_identity {
      type = "POSIX"

      posix_user {
        uid = 1001
        gid = 1001
      }
    }
  }
}
`, rName))
}

func testAccS3AccessPointAttachmentConfig_policy(rName string) string {
	return acctest.ConfigCompose(testAccS3AccessPointAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_s3_access_point_attachment" "test" {
  name = %[1]q
  type = "OPENZFS"

  openzfs_configuration {
    volume_id = aws_fsx_openzfs_volume.test.id

    file_system_identity {
      type = "POSIX"

      posix_user {
        uid = 1001
        gid = 1001

        secondary_gids = [1002, 1003]
      }
    }
  }

  s3_access_point {
    policy = jsonencode({
      Version = "2008-10-17"
      Statement = [{
        Effect = "Allow"
        Action = "s3:GetObjectTagging"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Resource = "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*"
      }]
    })
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName))
}

func testAccS3AccessPointAttachmentConfig_vpcConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccS3AccessPointAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_s3_access_point_attachment" "test" {
  name = %[1]q
  type = "OPENZFS"

  openzfs_configuration {
    volume_id = aws_fsx_openzfs_volume.test.id

    file_system_identity {
      type = "POSIX"

      posix_user {
        uid = 1001
        gid = 1001
      }
    }
  }

  s3_access_point {
    vpc_configuration {
      vpc_id = aws_vpc.test.id
    }
  }
}
`, rName))
}
