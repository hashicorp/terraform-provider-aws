// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdsql "github.com/hashicorp/terraform-provider-aws/internal/service/dsql"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v dsql.GetStreamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dsql_stream.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrClusterIdentifier), "aws_dsql_cluster.test", tfjsonpath.New(names.AttrIdentifier), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrFormat), tfknownvalue.StringExact(awstypes.StreamFormatJson)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ordering"), tfknownvalue.StringExact(awstypes.StreamOrderingUnordered)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("target_definition").AtSliceIndex(0).AtMapKey("kinesis").AtSliceIndex(0).AtMapKey(names.AttrRoleARN), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("target_definition").AtSliceIndex(0).AtMapKey("kinesis").AtSliceIndex(0).AtMapKey(names.AttrStreamARN), "aws_kinesis_stream.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "dsql", "cluster/{cluster_identifier}/stream/{stream_identifier}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.StringRegexp(regexache.MustCompile(acctest.RFC3339RegexPattern))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.StreamStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("stream_identifier"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, names.AttrClusterIdentifier, "stream_identifier"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrClusterIdentifier,
			},
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccDSQLStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v dsql.GetStreamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dsql_stream.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdsql.ResourceStream, resourceName),
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

func TestAccDSQLStream_replace_cluster_identifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v dsql.GetStreamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dsql_stream.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccStreamConfig_clusterIdentifierReplaced(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrClusterIdentifier), "aws_dsql_cluster.test2", tfjsonpath.New(names.AttrIdentifier), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccDSQLStream_replace_target_definition(t *testing.T) {
	ctx := acctest.Context(t)
	var v dsql.GetStreamOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dsql_stream.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccStreamConfig_targetDefinitionKinesisRoleARNReplaced(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStreamExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("target_definition").AtSliceIndex(0).AtMapKey("kinesis").AtSliceIndex(0).AtMapKey(names.AttrRoleARN), "aws_iam_role.test2", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccCheckStreamDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dsql_stream" {
				continue
			}

			clusterIdentifier := rs.Primary.Attributes[names.AttrClusterIdentifier]
			streamIdentifier := rs.Primary.Attributes["stream_identifier"]
			_, err := tfdsql.FindStreamByTwoPartKey(ctx, conn, clusterIdentifier, streamIdentifier)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Stream %s,%s still exists", clusterIdentifier, streamIdentifier)
		}

		return nil
	}
}

func testAccCheckStreamExists(ctx context.Context, t *testing.T, n string, v *dsql.GetStreamOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSQLClient(ctx)

		clusterIdentifier := rs.Primary.Attributes[names.AttrClusterIdentifier]
		streamIdentifier := rs.Primary.Attributes["stream_identifier"]
		output, err := tfdsql.FindStreamByTwoPartKey(ctx, conn, clusterIdentifier, streamIdentifier)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccStreamConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dsql_stream" "test" {
  cluster_identifier = aws_dsql_cluster.test.identifier
  format             = "JSON"
  ordering           = "UNORDERED"

  target_definition {
    kinesis {
      role_arn   = aws_iam_role.test.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_dsql_cluster" "test" {}

data "aws_service_principal" "test" {
  service_name = "dsql"
}

data "aws_iam_policy_document" "test_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.test.name]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role.json
}

data "aws_iam_policy_document" "test_role_policy" {
  statement {
    actions   = ["kinesis:DescribeStreamSummary", "kinesis:ListShards", "kinesis:PutRecord", "kinesis:PutRecords"]
    resources = [aws_kinesis_stream.test.arn]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test_role_policy.json
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccStreamConfig_clusterIdentifierReplaced(rName string) string {
	return fmt.Sprintf(`
resource "aws_dsql_stream" "test" {
  cluster_identifier = aws_dsql_cluster.test2.identifier
  format             = "JSON"
  ordering           = "UNORDERED"

  target_definition {
    kinesis {
      role_arn   = aws_iam_role.test.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_dsql_cluster" "test" {}

resource "aws_dsql_cluster" "test2" {}

data "aws_service_principal" "test" {
  service_name = "dsql"
}

data "aws_iam_policy_document" "test_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.test.name]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role.json
}

data "aws_iam_policy_document" "test_role_policy" {
  statement {
    actions   = ["kinesis:DescribeStreamSummary", "kinesis:ListShards", "kinesis:PutRecord", "kinesis:PutRecords"]
    resources = [aws_kinesis_stream.test.arn]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test_role_policy.json
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccStreamConfig_targetDefinitionKinesisRoleARNReplaced(rName string) string {
	return fmt.Sprintf(`
resource "aws_dsql_stream" "test" {
  cluster_identifier = aws_dsql_cluster.test.identifier
  format             = "JSON"
  ordering           = "UNORDERED"

  target_definition {
    kinesis {
      role_arn   = aws_iam_role.test2.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }
}

resource "aws_dsql_cluster" "test" {}

data "aws_service_principal" "test" {
  service_name = "dsql"
}

data "aws_iam_policy_document" "test_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.test.name]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role.json
}

data "aws_iam_policy_document" "test_role_policy" {
  statement {
    actions   = ["kinesis:DescribeStreamSummary", "kinesis:ListShards", "kinesis:PutRecord", "kinesis:PutRecords"]
    resources = [aws_kinesis_stream.test.arn]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test_role_policy.json
}

data "aws_service_principal" "test2" {
  service_name = "dsql"
}

data "aws_iam_policy_document" "test2_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = [data.aws_service_principal.test2.name]
    }
  }
}

resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  assume_role_policy = data.aws_iam_policy_document.test2_assume_role.json
}

data "aws_iam_policy_document" "test2_role_policy" {
  statement {
    actions   = ["kinesis:DescribeStreamSummary", "kinesis:ListShards", "kinesis:PutRecord", "kinesis:PutRecords"]
    resources = [aws_kinesis_stream.test.arn]
  }
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test2.id
  policy = data.aws_iam_policy_document.test2_role_policy.json
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}
