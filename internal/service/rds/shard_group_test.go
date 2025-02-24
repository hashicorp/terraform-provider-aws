// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSShardGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBShardGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_shard_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckShardGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccShardGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("rds", regexache.MustCompile(`shard-group:shardgroup-[a-z0-9]+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("compute_redundancy"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_shard_group_identifier"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_shard_group_resource_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_acu"), knownvalue.Float64Exact(1000)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_acu"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPubliclyAccessible), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccShardGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "db_shard_group_identifier",
			},
		},
	})
}

func TestAccRDSShardGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBShardGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_shard_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckShardGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccShardGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrds.ResourceShardGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSShardGroup_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBShardGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_shard_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckShardGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccShardGroupConfig_full(rName, 1200, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("rds", regexache.MustCompile(`shard-group:shardgroup-[a-z0-9]+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("compute_redundancy"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_shard_group_identifier"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_shard_group_resource_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_acu"), knownvalue.Float64Exact(1200)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_acu"), knownvalue.Float64Exact(120)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPubliclyAccessible), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccShardGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "db_shard_group_identifier",
			},
			{
				Config: testAccShardGroupConfig_full(rName, 1300, 130),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("rds", regexache.MustCompile(`shard-group:shardgroup-[a-z0-9]+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("compute_redundancy"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_shard_group_identifier"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("db_shard_group_resource_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEndpoint), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_acu"), knownvalue.Float64Exact(1300)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_acu"), knownvalue.Float64Exact(130)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPubliclyAccessible), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccRDSShardGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBShardGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rds_shard_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckShardGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccShardGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccShardGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "db_shard_group_identifier",
			},
			{
				Config: testAccShardGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccShardGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckShardGroupExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckShardGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_shard_group" {
				continue
			}

			_, err := tfrds.FindDBShardGroupByID(ctx, conn, rs.Primary.Attributes["db_shard_group_identifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Shard Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckShardGroupExists(ctx context.Context, n string, v *awstypes.DBShardGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBShardGroupByID(ctx, conn, rs.Primary.Attributes["db_shard_group_identifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccShardGroupImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["db_shard_group_identifier"], nil
	}
}

func testAccShardGroupConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

# https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/limitless-create-cluster.html.
resource "aws_rds_cluster" "test" {
  cluster_identifier                    = %[1]q
  engine                                = "aurora-postgresql"
  engine_version                        = "16.6-limitless"
  engine_mode                           = ""
  storage_type                          = "aurora-iopt1"
  cluster_scalability_type              = "limitless"
  master_username                       = "tfacctest"
  master_password                       = "avoid-plaintext-passwords"
  skip_final_snapshot                   = true
  performance_insights_enabled          = true
  performance_insights_retention_period = 31
  enabled_cloudwatch_logs_exports       = ["postgresql"]
  monitoring_interval                   = 5
  monitoring_role_arn                   = aws_iam_role.test.arn
}
`, rName)
}

func testAccShardGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccShardGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_shard_group" "test" {
  db_shard_group_identifier = %[1]q
  db_cluster_identifier     = aws_rds_cluster.test.id
  max_acu                   = 1000
}
`, rName))
}

func testAccShardGroupConfig_full(rName string, maxACU, minACU int) string {
	return acctest.ConfigCompose(testAccShardGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_shard_group" "test" {
  db_shard_group_identifier = %[1]q
  db_cluster_identifier     = aws_rds_cluster.test.id
  max_acu                   = %[2]d
  min_acu                   = %[3]d
  publicly_accessible       = true
  compute_redundancy        = 1
}
`, rName, maxACU, minACU))
}

func testAccShardGroupConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccShardGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_shard_group" "test" {
  db_shard_group_identifier = %[1]q
  db_cluster_identifier     = aws_rds_cluster.test.id
  max_acu                   = 1000

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccShardGroupConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccShardGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_shard_group" "test" {
  db_shard_group_identifier = %[1]q
  db_cluster_identifier     = aws_rds_cluster.test.id
  max_acu                   = 1000

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
