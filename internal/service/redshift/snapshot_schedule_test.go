// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftSnapshotSchedule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotSchedule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "redshift", "snapshotschedule:{id}"),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "definitions.*", "rate(12 hours)"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, rName),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotSchedule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshift.ResourceSnapshotSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_identifierGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotSchedule
	resourceName := "aws_redshift_snapshot_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_identifierGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", sdkid.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_identifierPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotSchedule
	resourceName := "aws_redshift_snapshot_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_identifierPrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrIdentifier, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "identifier_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotSchedule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_multipleDefinitions(rName, "cron(30 12 *)", "cron(15 6 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "definitions.*", "cron(30 12 *)"),
					resource.TestCheckTypeSetElemAttr(resourceName, "definitions.*", "cron(15 6 *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
				},
			},
			{
				Config: testAccSnapshotScheduleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "definitions.*", "rate(12 hours)"),
				),
			},
			{
				Config: testAccSnapshotScheduleConfig_multipleDefinitions(rName, "cron(30 8 *)", "cron(15 10 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "definitions.*", "cron(30 8 *)"),
					resource.TestCheckTypeSetElemAttr(resourceName, "definitions.*", "cron(15 10 *)"),
				),
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_withDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SnapshotSchedule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test Schedule"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_withForceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var snapshotSchedule awstypes.SnapshotSchedule
	var cluster awstypes.Cluster
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.test"
	clusterResourceName := "aws_redshift_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSnapshotScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(ctx, t, resourceName, &snapshotSchedule),
					testAccCheckClusterExists(ctx, t, clusterResourceName, &cluster),
					testAccCheckSnapshotScheduleCreateSnapshotScheduleAssociation(ctx, t, &cluster, &snapshotSchedule),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrForceDestroy,
				},
			},
		},
	})
}

func testAccCheckSnapshotScheduleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_snapshot_schedule" {
				continue
			}

			_, err := tfredshift.FindSnapshotScheduleByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Snapshot Schedule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSnapshotScheduleExists(ctx context.Context, t *testing.T, n string, v *awstypes.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		output, err := tfredshift.FindSnapshotScheduleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSnapshotScheduleCreateSnapshotScheduleAssociation(ctx context.Context, t *testing.T, cluster *awstypes.Cluster, snapshotSchedule *awstypes.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		_, err := conn.ModifyClusterSnapshotSchedule(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    cluster.ClusterIdentifier,
			ScheduleIdentifier:   snapshotSchedule.ScheduleIdentifier,
			DisassociateSchedule: aws.Bool(false),
		})

		if err != nil {
			return err
		}

		if _, err := tfredshift.WaitSnapshotScheduleAssociationCreated(ctx, conn, aws.ToString(cluster.ClusterIdentifier), aws.ToString(snapshotSchedule.ScheduleIdentifier)); err != nil {
			return err
		}

		return nil
	}
}

func testAccSnapshotScheduleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "test" {
  identifier = %[1]q
  definitions = [
    "rate(12 hours)",
  ]
}
`, rName)
}

func testAccSnapshotScheduleConfig_identifierGenerated() string {
	return `
resource "aws_redshift_snapshot_schedule" "test" {
  definitions = [
    "rate(12 hours)",
  ]
}
`
}

func testAccSnapshotScheduleConfig_identifierPrefix(prefix string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "test" {
  identifier_prefix = %[1]q
  definitions = [
    "rate(12 hours)",
  ]
}
`, prefix)
}

func testAccSnapshotScheduleConfig_multipleDefinitions(rName, definition1, definition2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "test" {
  identifier = %[1]q
  definitions = [
    %[2]q,
    %[3]q,
  ]
}
`, rName, definition1, definition2)
}

func testAccSnapshotScheduleConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "test" {
  identifier  = %[1]q
  description = "Test Schedule"
  definitions = [
    "rate(12 hours)",
  ]
}
`, rName)
}

func testAccSnapshotScheduleConfig_forceDestroy(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "test" {
  identifier = %[1]q
  definitions = [
    "rate(12 hours)",
  ]
  force_destroy = true
}
`, rName))
}
