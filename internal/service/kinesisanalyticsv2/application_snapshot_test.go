// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkinesisanalyticsv2 "github.com/hashicorp/terraform-provider-aws/internal/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisAnalyticsV2ApplicationSnapshot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationSnapshotExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "application_name", applicationResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "application_version_id", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
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

func TestAccKinesisAnalyticsV2ApplicationSnapshot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationSnapshotExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkinesisanalyticsv2.ResourceApplicationSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisAnalyticsV2ApplicationSnapshot_Disappears_application(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisAnalyticsV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationSnapshotExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkinesisanalyticsv2.ResourceApplication(), applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationSnapshotDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KinesisAnalyticsV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesisanalyticsv2_application_snapshot" {
				continue
			}

			_, err := tfkinesisanalyticsv2.FindSnapshotDetailsByTwoPartKey(ctx, conn, rs.Primary.Attributes["application_name"], rs.Primary.Attributes["snapshot_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Analytics v2 Application Snapshot %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckApplicationSnapshotExists(ctx context.Context, t *testing.T, n string, v *awstypes.SnapshotDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KinesisAnalyticsV2Client(ctx)

		output, err := tfkinesisanalyticsv2.FindSnapshotDetailsByTwoPartKey(ctx, conn, rs.Primary.Attributes["application_name"], rs.Primary.Attributes["snapshot_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationSnapshotConfig_basic(rName string) string {
	return testAccApplicationConfig_startSnapshotableFlink(rName, "SKIP_RESTORE_FROM_SNAPSHOT", "", false)
}
